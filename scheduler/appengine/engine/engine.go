// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package engine implements the core logic of the scheduler service.
package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/api/pubsub/v1"

	ds "github.com/luci/gae/service/datastore"
	"github.com/luci/gae/service/info"
	mc "github.com/luci/gae/service/memcache"
	tq "github.com/luci/gae/service/taskqueue"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/data/rand/mathrand"
	"github.com/luci/luci-go/common/data/stringset"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/server/auth"
	"github.com/luci/luci-go/server/auth/identity"
	"github.com/luci/luci-go/server/auth/signing"
	"github.com/luci/luci-go/server/tokens"

	"github.com/luci/luci-go/scheduler/appengine/catalog"
	"github.com/luci/luci-go/scheduler/appengine/schedule"
	"github.com/luci/luci-go/scheduler/appengine/task"
)

// Engine manages all scheduler jobs: keeps track of their state, runs state
// machine transactions, starts new invocations, etc. A method returns
// errors.Transient if the error is non-fatal and the call should be retried
// later. Any other error means that retry won't help.
type Engine interface {
	// GetAllProjects returns a list of all projects that have at least one
	// enabled scheduler job.
	GetAllProjects(c context.Context) ([]string, error)

	// GetAllJobs returns a list of all enabled scheduler jobs in no particular
	// order.
	GetAllJobs(c context.Context) ([]*Job, error)

	// GetProjectJobs returns a list of enabled scheduler jobs of some project
	// in no particular order.
	GetProjectJobs(c context.Context, projectID string) ([]*Job, error)

	// GetJob returns single scheduler job given its full ID or nil if no such
	// job.
	GetJob(c context.Context, jobID string) (*Job, error)

	// ListInvocations returns invocations of a job, most recent first.
	// Returns fetched invocations and cursor string if there's more.
	ListInvocations(c context.Context, jobID string, pageSize int, cursor string) ([]*Invocation, string, error)

	// GetInvocation returns single invocation of some job given its ID.
	GetInvocation(c context.Context, jobID string, invID int64) (*Invocation, error)

	// GetInvocationsByNonce returns a list of Invocations with given nonce.
	//
	// Invocation nonce is a random number that identifies an intent to start
	// an invocation. Normally one nonce corresponds to one Invocation entity,
	// but there can be more if job fails to start with a transient error.
	GetInvocationsByNonce(c context.Context, invNonce int64) ([]*Invocation, error)

	// UpdateProjectJobs adds new, removes old and updates existing jobs.
	UpdateProjectJobs(c context.Context, projectID string, defs []catalog.Definition) error

	// ResetAllJobsOnDevServer forcefully resets state of all enabled jobs.
	// Supposed to be used only on devserver, where task queue stub state is not
	// preserved between appserver restarts and it messes everything.
	ResetAllJobsOnDevServer(c context.Context) error

	// ExecuteSerializedAction is called via a task queue to execute an action
	// produced by job state machine transition. These actions are POSTed
	// to TimersQueue and InvocationsQueue defined in Config by Engine.
	// 'retryCount' is 0 on first attempt, 1 if task queue service retries
	// request once, 2 - if twice, and so on. Returning transient errors here
	// causes the task queue to retry the task.
	ExecuteSerializedAction(c context.Context, body []byte, retryCount int) error

	// ProcessPubSubPush is called whenever incoming PubSub message is received.
	ProcessPubSubPush(c context.Context, body []byte) error

	// PullPubSubOnDevServer is called on dev server to pull messages from PubSub
	// subscription associated with given publisher.
	//
	// It is needed to be able to manually tests PubSub related workflows on dev
	// server, since dev server can't accept PubSub push messages.
	PullPubSubOnDevServer(c context.Context, taskManagerName, publisher string) error

	// TriggerInvocation launches job invocation right now if job isn't running
	// now. Used by "Run now" UI button.
	//
	// Returns new invocation nonce (a random number that identifies an intent to
	// start an invocation). Normally one nonce corresponds to one Invocation
	// entity, but there can be more if job fails to start with a transient error.
	TriggerInvocation(c context.Context, jobID string, triggeredBy identity.Identity) (int64, error)

	// PauseJob replaces job's schedule with "triggered", effectively preventing
	// it from running automatically (until it is resumed). Manual invocations are
	// still allowed. Does nothing if job is already paused. Any pending or
	// running invocations are still executed.
	PauseJob(c context.Context, jobID string, who identity.Identity) error

	// ResumeJob resumed paused job. Doesn't nothing if the job is not paused.
	ResumeJob(c context.Context, jobID string, who identity.Identity) error

	// AbortInvocation forcefully moves the invocation to failed state.
	//
	// It opportunistically tries to send "abort" signal to a job runner if it
	// supports cancellation, but it doesn't wait for reply. It proceeds to
	// modifying local state in the scheduler service datastore immediately.
	//
	// AbortInvocation can be used to manually "unstuck" jobs that got stuck due
	// to missing PubSub notifications or other kinds of unexpected conditions.
	//
	// Does nothing if invocation is already in some final state.
	AbortInvocation(c context.Context, jobID string, invID int64, who identity.Identity) error

	// AbortJob resets the job to scheduled state, aborting a currently pending or
	// running invocation (if any).
	//
	// Returns nil if the job is not currently running.
	AbortJob(c context.Context, jobID string, who identity.Identity) error
}

// Config contains parameters for the engine.
type Config struct {
	Catalog              catalog.Catalog // provides task.Manager's to run tasks
	TimersQueuePath      string          // URL of a task queue handler for timer ticks
	TimersQueueName      string          // queue name for timer ticks
	InvocationsQueuePath string          // URL of a task queue handler that starts jobs
	InvocationsQueueName string          // queue name for job starts
	PubSubPushPath       string          // URL to use in PubSub push config
}

// NewEngine returns default implementation of Engine.
func NewEngine(conf Config) Engine {
	return &engineImpl{
		Config:    conf,
		doneFlags: make(map[string]bool),
	}
}

//// Implementation.

const (
	// invocationRetryLimit is how many times to retry an invocation before giving
	// up and resuming the job's schedule.
	invocationRetryLimit = 5

	// maxInvocationRetryBackoff is how long to wait before retrying a failed
	// invocation.
	maxInvocationRetryBackoff = 10 * time.Second

	// debugLogSizeLimit is how many bytes the invocation debug log can be before
	// it gets trimmed. See 'trimDebugLog'. The debug log isn't supposed to be
	// huge.
	debugLogSizeLimit = 200000

	// debugLogTailLines is how many last log lines to keep when trimming the log.
	debugLogTailLines = 100
)

// actionTaskPayload is payload for task queue jobs emitted by the engine.
//
// Serialized as JSON, produced by enqueueJobActions and enqueueInvTimers, used
// as inputs in ExecuteSerializedAction.
//
// Union of all possible payloads for simplicity.
type actionTaskPayload struct {
	JobID string `json:",omitempty"` // ID of relevant Job
	InvID int64  `json:",omitempty"` // ID of relevant Invocation

	// For Job actions and timers (InvID == 0).
	Kind                string `json:",omitempty"` // defines what fields below to examine
	TickNonce           int64  `json:",omitempty"` // valid for "TickLaterAction" kind
	InvocationNonce     int64  `json:",omitempty"` // valid for "StartInvocationAction" kind
	TriggeredBy         string `json:",omitempty"` // valid for "StartInvocationAction" kind
	Overruns            int    `json:",omitempty"` // valid for "RecordOverrunAction" kind
	RunningInvocationID int64  `json:",omitempty"` // valid for "RecordOverrunAction" kind

	// For Invocation actions and timers (InvID != 0).
	InvTimer *invocationTimer `json:",omitempty"` // used for AddTimer calls
}

// invocationTimer is carried as part of task queue task payload for tasks
// created by AddTimer calls.
//
// It will be serialized to JSON, so all fields are public.
type invocationTimer struct {
	Delay   time.Duration
	Name    string
	Payload []byte
}

// Job stores the last known definition of a scheduler job, as well as its
// current state. Root entity, its kind is "Job".
type Job struct {
	_kind  string         `gae:"$kind,Job"`
	_extra ds.PropertyMap `gae:"-,extra"`

	// cachedSchedule and cachedScheduleErr are used by parseSchedule().
	cachedSchedule    *schedule.Schedule `gae:"-"`
	cachedScheduleErr error              `gae:"-"`

	// JobID is '<ProjectID>/<JobName>' string. JobName is unique with a project,
	// but not globally. JobID is unique globally.
	JobID string `gae:"$id"`

	// ProjectID exists for indexing. It matches <projectID> portion of JobID.
	ProjectID string

	// Flavor describes what category of jobs this is, see the enum.
	Flavor catalog.JobFlavor `gae:",noindex"`

	// Enabled is false if the job was disabled or removed from config.
	//
	// Disabled jobs do not show up in UI at all (they are still kept in the
	// datastore though, for audit purposes).
	Enabled bool

	// Paused is true if job's schedule is ignored and job can only be started
	// manually via "Run now" button.
	Paused bool `gae:",noindex"`

	// Revision is last seen job definition revision.
	Revision string `gae:",noindex"`

	// RevisionURL is URL to human readable page with config file at
	// an appropriate revision.
	RevisionURL string `gae:",noindex"`

	// Schedule is the job's schedule in regular cron expression format.
	Schedule string `gae:",noindex"`

	// Task is the job's payload in serialized form. Opaque from the point of view
	// of the engine. See Catalog.UnmarshalTask().
	Task []byte `gae:",noindex"`

	// State is the job's state machine state, see StateMachine.
	State JobState
}

// effectiveSchedule returns schedule string to use for the job, considering its
// Paused field.
//
// Paused jobs always use "triggered" schedule.
func (e *Job) effectiveSchedule() string {
	if e.Paused {
		return "triggered"
	}
	return e.Schedule
}

// parseSchedule returns *Schedule object, parsing e.Schedule field.
// If job is paused e.Schedule field is ignored and triggered schedule is
// returned instead.
func (e *Job) parseSchedule() (*schedule.Schedule, error) {
	if e.cachedSchedule == nil && e.cachedScheduleErr == nil {
		hash := fnv.New64()
		hash.Write([]byte(e.JobID))
		seed := hash.Sum64()
		e.cachedSchedule, e.cachedScheduleErr = schedule.Parse(e.effectiveSchedule(), seed)
		if e.cachedSchedule == nil && e.cachedScheduleErr == nil {
			panic("no schedule and no error")
		}
	}
	return e.cachedSchedule, e.cachedScheduleErr
}

// isEqual returns true iff 'e' is equal to 'other'.
func (e *Job) isEqual(other *Job) bool {
	return e == other || (e.JobID == other.JobID &&
		e.ProjectID == other.ProjectID &&
		e.Flavor == other.Flavor &&
		e.Enabled == other.Enabled &&
		e.Paused == other.Paused &&
		e.Revision == other.Revision &&
		e.RevisionURL == other.RevisionURL &&
		e.Schedule == other.Schedule &&
		bytes.Equal(e.Task, other.Task) &&
		e.State == other.State)
}

// matches returns true if job definition in the entity matches the one
// specified by catalog.Definition struct. UpdateProjectJobs skips updates for
// such jobs (assuming they are up-to-date).
func (e *Job) matches(def catalog.Definition) bool {
	return e.JobID == def.JobID &&
		e.Flavor == def.Flavor &&
		e.Schedule == def.Schedule &&
		bytes.Equal(e.Task, def.Task)
}

// Invocation entity stores single attempt to run a job. Its parent entity
// is corresponding Job, its ID is generated based on time.
type Invocation struct {
	_kind  string         `gae:"$kind,Invocation"`
	_extra ds.PropertyMap `gae:"-,extra"`

	// ID is identifier of this particular attempt to run a job. Multiple attempts
	// to start an invocation result in multiple entities with different IDs, but
	// with same InvocationNonce.
	ID int64 `gae:"$id"`

	// JobKey is the key of parent Job entity.
	JobKey *ds.Key `gae:"$parent"`

	// Started is time when this invocation was created.
	Started time.Time `gae:",noindex"`

	// Finished is time when this invocation transitioned to a terminal state.
	Finished time.Time `gae:",noindex"`

	// InvocationNonce identifies a request to start a job, produced by
	// StateMachine.
	InvocationNonce int64

	// TriggeredBy is identity of whoever triggered the invocation, if it was
	// triggered via TriggerInvocation ("Run now" button).
	//
	// Empty identity string if it was triggered by the service itself.
	TriggeredBy identity.Identity

	// Revision is revision number of config.cfg when this invocation was created.
	// For informational purpose.
	Revision string `gae:",noindex"`

	// RevisionURL is URL to human readable page with config file at
	// an appropriate revision. For informational purpose.
	RevisionURL string `gae:",noindex"`

	// Task is the job payload for this invocation in binary serialized form.
	// For informational purpose. See Catalog.UnmarshalTask().
	Task []byte `gae:",noindex"`

	// DebugLog is short free form text log with debug messages.
	DebugLog string `gae:",noindex"`

	// RetryCount is 0 on a first attempt to launch the task. Increased with each
	// retry. For informational purposes.
	RetryCount int64 `gae:",noindex"`

	// Status is current status of the invocation (e.g. "RUNNING"), see the enum.
	Status task.Status

	// ViewURL is optional URL to a human readable page with task status, e.g.
	// Swarming task page. Populated by corresponding TaskManager.
	ViewURL string `gae:",noindex"`

	// TaskData is a storage where TaskManager can keep task-specific state
	// between calls.
	TaskData []byte `gae:",noindex"`

	// MutationsCount is used for simple compare-and-swap transaction control.
	// It is incremented on each change to the entity. See 'saveImpl' below.
	MutationsCount int64 `gae:",noindex"`
}

// isEqual returns true iff 'e' is equal to 'other'.
func (e *Invocation) isEqual(other *Invocation) bool {
	return e == other || (e.ID == other.ID &&
		(e.JobKey == other.JobKey || e.JobKey.Equal(other.JobKey)) &&
		e.Started == other.Started &&
		e.Finished == other.Finished &&
		e.InvocationNonce == other.InvocationNonce &&
		e.Revision == other.Revision &&
		e.RevisionURL == other.RevisionURL &&
		bytes.Equal(e.Task, other.Task) &&
		e.DebugLog == other.DebugLog &&
		e.RetryCount == other.RetryCount &&
		e.Status == other.Status &&
		e.ViewURL == other.ViewURL &&
		bytes.Equal(e.TaskData, other.TaskData) &&
		e.MutationsCount == other.MutationsCount)
}

// debugLog appends a line to DebugLog field.
func (e *Invocation) debugLog(c context.Context, format string, args ...interface{}) {
	debugLog(c, &e.DebugLog, format, args...)
}

// trimDebugLog makes sure DebugLog field doesn't exceed limits.
//
// It cuts the middle of the log. We need to do this to keep the entity small
// enough to fit the datastore limits.
func (e *Invocation) trimDebugLog() {
	if len(e.DebugLog) <= debugLogSizeLimit {
		return
	}

	const cutMsg = "--- the log has been cut here ---"
	giveUp := func() {
		e.DebugLog = e.DebugLog[:debugLogSizeLimit-len(cutMsg)-2] + "\n" + cutMsg + "\n"
	}

	// We take last debugLogTailLines lines of log and move them "up", so that
	// the total log size is less than debugLogSizeLimit. We then put a line with
	// the message that some log lines have been cut. If these operations are not
	// possible (e.g. we have some giant lines or something), we give up and just
	// cut the end of the log.

	// Find debugLogTailLines-th "\n" from the end, e.DebugLog[tailStart:] is the
	// log tail.
	tailStart := len(e.DebugLog)
	for i := 0; i < debugLogTailLines; i++ {
		tailStart = strings.LastIndex(e.DebugLog[:tailStart-1], "\n")
		if tailStart <= 0 {
			giveUp()
			return
		}
	}
	tailStart++

	// Figure out how many bytes of head we can keep to make trimmed log small
	// enough.
	tailLen := len(e.DebugLog) - tailStart + len(cutMsg) + 1
	headSize := debugLogSizeLimit - tailLen
	if headSize <= 0 {
		giveUp()
		return
	}

	// Find last "\n" in the head.
	headEnd := strings.LastIndex(e.DebugLog[:headSize], "\n")
	if headEnd <= 0 {
		giveUp()
		return
	}

	// We want to keep 50 lines of the head no matter what.
	headLines := strings.Count(e.DebugLog[:headEnd], "\n")
	if headLines < 50 {
		giveUp()
		return
	}

	// Remove duplicated 'cutMsg' lines. They may appear if 'debugLog' (followed
	// by 'trimDebugLog') is called on already trimmed log multiple times.
	lines := strings.Split(e.DebugLog[:headEnd], "\n")
	lines = append(lines, cutMsg)
	lines = append(lines, strings.Split(e.DebugLog[tailStart:], "\n")...)
	trimmed := make([]byte, 0, debugLogSizeLimit)
	trimmed = append(trimmed, lines[0]...)
	for i := 1; i < len(lines); i++ {
		if !(lines[i-1] == cutMsg && lines[i] == cutMsg) {
			trimmed = append(trimmed, '\n')
			trimmed = append(trimmed, lines[i]...)
		}
	}
	e.DebugLog = string(trimmed)
}

// Jan 1 2015, in UTC.
var invocationIDEpoch time.Time

func init() {
	var err error
	invocationIDEpoch, err = time.Parse(time.RFC822, "01 Jan 15 00:00 UTC")
	if err != nil {
		panic(err)
	}
}

// generateInvocationID is called within a transaction to pick a new Invocation
// ID and ensure it isn't taken yet.
//
// Format of the invocation ID:
//   - highest order bit set to 0 to keep the value positive.
//   - next 43 bits set to negated time since some predefined epoch, in ms.
//   - next 16 bits are generated by math.Rand
//   - next 4 bits set to 0. They indicate ID format.
func generateInvocationID(c context.Context, parent *ds.Key) (int64, error) {
	rnd := mathrand.Get(c)

	// See http://play.golang.org/p/POpQzpT4Up.
	invTs := int64(clock.Now(c).UTC().Sub(invocationIDEpoch) / time.Millisecond)
	invTs = ^invTs & 8796093022207 // 0b111....1, 42 bits (clear highest bit)
	invTs = invTs << 20

	for i := 0; i < 10; i++ {
		randSuffix := rnd.Int63n(65536)
		invID := invTs | (randSuffix << 4)
		exists, err := ds.Exists(c, ds.NewKey(c, "Invocation", "", invID, parent))
		if err != nil {
			return 0, err
		}
		if !exists.All() {
			return invID, nil
		}
	}

	return 0, errors.New("could not find available invocationID after 10 attempts")
}

// debugLog mutates a string by appending a line to it.
func debugLog(c context.Context, str *string, format string, args ...interface{}) {
	prefix := clock.Now(c).UTC().Format("[15:04:05.000] ")
	*str += prefix + fmt.Sprintf(format+"\n", args...)
}

////

type engineImpl struct {
	Config

	lock      sync.Mutex
	doneFlags map[string]bool // see doIfNotDone

	// configureTopic is used by prepareTopic, mocked in tests.
	configureTopic func(c context.Context, topic, sub, pushURL, publisher string) error
}

// doIfNotDone calls callback only if it wasn't called before.
//
// Works on best effort basis: callback can and will be called multiple times
// (just not the every time 'doIfNotDone' is called).
//
// Keeps "done" flag in local memory and in memcache (using 'key' as
// identifier). The callback should be idempotent, since it still may be called
// multiple times if multiple processes attempt to execute the action at once.
func (e *engineImpl) doIfNotDone(c context.Context, key string, cb func() error) error {
	// Check the local cache.
	e.lock.Lock()
	if e.doneFlags[key] {
		e.lock.Unlock()
		return nil
	}
	e.lock.Unlock()

	// Check the global cache.
	switch _, err := mc.GetKey(c, key); {
	case err == nil:
		e.lock.Lock()
		defer e.lock.Unlock()
		e.doneFlags[key] = true
		return nil
	case err == mc.ErrCacheMiss:
		break
	default:
		return errors.WrapTransient(err)
	}

	// Do it.
	if err := cb(); err != nil {
		return err
	}

	// Store in the global cache. Ignore errors, it's not a big deal.
	item := mc.NewItem(c, key)
	item.SetValue([]byte("ok"))
	item.SetExpiration(24 * time.Hour)
	if err := mc.Set(c, item); err != nil {
		logging.Warningf(c, "Failed to write item to memcache - %s", err)
	}

	// Store in the local cache.
	e.lock.Lock()
	defer e.lock.Unlock()
	e.doneFlags[key] = true
	return nil
}

func (e *engineImpl) GetAllProjects(c context.Context) ([]string, error) {
	q := ds.NewQuery("Job").
		Eq("Enabled", true).
		Project("ProjectID").
		Distinct(true)
	entities := []Job{}
	if err := ds.GetAll(c, q, &entities); err != nil {
		return nil, errors.WrapTransient(err)
	}
	// Filter out duplicates, sort.
	projects := stringset.New(len(entities))
	for _, ent := range entities {
		projects.Add(ent.ProjectID)
	}
	out := projects.ToSlice()
	sort.Strings(out)
	return out, nil
}

func (e *engineImpl) GetAllJobs(c context.Context) ([]*Job, error) {
	q := ds.NewQuery("Job").Eq("Enabled", true)
	return e.queryEnabledJobs(c, q)
}

func (e *engineImpl) GetProjectJobs(c context.Context, projectID string) ([]*Job, error) {
	q := ds.NewQuery("Job").Eq("Enabled", true).Eq("ProjectID", projectID)
	return e.queryEnabledJobs(c, q)
}

func (e *engineImpl) queryEnabledJobs(c context.Context, q *ds.Query) ([]*Job, error) {
	entities := []*Job{}
	if err := ds.GetAll(c, q, &entities); err != nil {
		return nil, errors.WrapTransient(err)
	}
	// Non-ancestor query used, need to recheck filters.
	filtered := make([]*Job, 0, len(entities))
	for _, job := range entities {
		if job.Enabled {
			filtered = append(filtered, job)
		}
	}
	return filtered, nil
}

func (e *engineImpl) GetJob(c context.Context, jobID string) (*Job, error) {
	job := &Job{JobID: jobID}
	switch err := ds.Get(c, job); {
	case err == nil:
		return job, nil
	case err == ds.ErrNoSuchEntity:
		return nil, nil
	default:
		return nil, errors.WrapTransient(err)
	}
}

func (e *engineImpl) ListInvocations(c context.Context, jobID string, pageSize int, cursor string) ([]*Invocation, string, error) {
	if pageSize == 0 || pageSize > 500 {
		pageSize = 500
	}

	// Deserialize the cursor.
	var cursorObj ds.Cursor
	if cursor != "" {
		var err error
		cursorObj, err = ds.DecodeCursor(c, cursor)
		if err != nil {
			return nil, "", err
		}
	}

	// Prepare the query. Fetch 'pageSize' worth of entities as a single batch.
	q := ds.NewQuery("Invocation").
		Ancestor(ds.NewKey(c, "Job", jobID, 0, nil)).
		Order("__key__")
	q.Limit(int32(pageSize))
	if cursorObj != nil {
		q = q.Start(cursorObj)
	}

	// Fetch pageSize worth of invocations, then grab the cursor.
	out := make([]*Invocation, 0, pageSize)
	var newCursor string
	err := ds.Run(c, q, func(obj *Invocation, getCursor ds.CursorCB) error {
		out = append(out, obj)
		if len(out) < pageSize {
			return nil
		}
		c, err := getCursor()
		if err != nil {
			return err
		}
		newCursor = c.String()
		return ds.Stop
	})
	if err != nil {
		return nil, "", errors.WrapTransient(err)
	}
	return out, newCursor, nil
}

func (e *engineImpl) GetInvocation(c context.Context, jobID string, invID int64) (*Invocation, error) {
	inv := &Invocation{
		ID:     invID,
		JobKey: ds.NewKey(c, "Job", jobID, 0, nil),
	}
	switch err := ds.Get(c, inv); {
	case err == nil:
		return inv, nil
	case err == ds.ErrNoSuchEntity:
		return nil, nil
	default:
		return nil, errors.WrapTransient(err)
	}
}

func (e *engineImpl) GetInvocationsByNonce(c context.Context, invNonce int64) ([]*Invocation, error) {
	q := ds.NewQuery("Invocation").Eq("InvocationNonce", invNonce)
	entities := []*Invocation{}
	if err := ds.GetAll(c, q, &entities); err != nil {
		return nil, errors.WrapTransient(err)
	}
	return entities, nil
}

func (e *engineImpl) UpdateProjectJobs(c context.Context, projectID string, defs []catalog.Definition) error {
	// JobID -> *Job map.
	existing, err := e.getProjectJobs(c, projectID)
	if err != nil {
		return err
	}
	// JobID -> new definition revision map.
	updated := make(map[string]string, len(defs))
	for _, def := range defs {
		updated[def.JobID] = def.Revision
	}
	// List of job ids to disable.
	toDisable := []string{}
	for id := range existing {
		if updated[id] == "" {
			toDisable = append(toDisable, id)
		}
	}

	wg := sync.WaitGroup{}

	// Add new jobs, update existing ones.
	updateErrs := errors.NewLazyMultiError(len(defs))
	for i, def := range defs {
		if ent := existing[def.JobID]; ent != nil {
			if ent.Enabled && ent.matches(def) {
				continue
			}
		}
		wg.Add(1)
		go func(i int, def catalog.Definition) {
			updateErrs.Assign(i, e.updateJob(c, def))
			wg.Done()
		}(i, def)
	}

	// Disable old jobs.
	disableErrs := errors.NewLazyMultiError(len(toDisable))
	for i, jobID := range toDisable {
		wg.Add(1)
		go func(i int, jobID string) {
			disableErrs.Assign(i, e.disableJob(c, jobID))
			wg.Done()
		}(i, jobID)
	}

	wg.Wait()
	if updateErrs.Get() == nil && disableErrs.Get() == nil {
		return nil
	}
	return errors.WrapTransient(errors.NewMultiError(updateErrs.Get(), disableErrs.Get()))
}

func (e *engineImpl) ResetAllJobsOnDevServer(c context.Context) error {
	if !info.IsDevAppServer(c) {
		return errors.New("ResetAllJobsOnDevServer must not be used in production")
	}
	q := ds.NewQuery("Job").Eq("Enabled", true)
	keys := []*ds.Key{}
	if err := ds.GetAll(c, q, &keys); err != nil {
		return errors.WrapTransient(err)
	}
	wg := sync.WaitGroup{}
	errs := errors.NewLazyMultiError(len(keys))
	for i, key := range keys {
		wg.Add(1)
		go func(i int, key *ds.Key) {
			errs.Assign(i, e.resetJobOnDevServer(c, key.StringID()))
			wg.Done()
		}(i, key)
	}
	wg.Wait()
	return errors.WrapTransient(errs.Get())
}

// getProjectJobs fetches from ds all enabled jobs belonging to a given
// project.
func (e *engineImpl) getProjectJobs(c context.Context, projectID string) (map[string]*Job, error) {
	q := ds.NewQuery("Job").
		Eq("Enabled", true).
		Eq("ProjectID", projectID)
	entities := []*Job{}
	if err := ds.GetAll(c, q, &entities); err != nil {
		return nil, errors.WrapTransient(err)
	}
	out := make(map[string]*Job, len(entities))
	for _, job := range entities {
		if job.Enabled && job.ProjectID == projectID {
			out[job.JobID] = job
		}
	}
	return out, nil
}

// txnCallback is passed to 'txn' and it modifies 'job' in place. 'txn' then
// puts it into datastore. The callback may return errSkipPut to instruct 'txn'
// not to call datastore 'Put'. The callback may do other transactional things
// using the context.
type txnCallback func(c context.Context, job *Job, isNew bool) error

// errSkipPut can be returned by txnCallback to cancel ds.Put call.
var errSkipPut = errors.New("errSkipPut")

// defaultTransactionOptions is used for all transactions. The scheduler service
// has no user facing API, all activity is in background task queues. So tune it
// to do more retries.
var defaultTransactionOptions = ds.TransactionOptions{
	Attempts: 10,
}

// txn reads Job, calls callback, then dumps the modified entity back into
// datastore (unless callback returns errSkipPut).
func (e *engineImpl) txn(c context.Context, jobID string, txn txnCallback) error {
	c = logging.SetField(c, "JobID", jobID)
	fatal := false
	attempt := 0
	err := ds.RunInTransaction(c, func(c context.Context) error {
		attempt++
		if attempt != 1 {
			logging.Warningf(c, "Retrying transaction...")
		}
		stored := Job{JobID: jobID}
		err := ds.Get(c, &stored)
		if err != nil && err != ds.ErrNoSuchEntity {
			return err
		}
		modified := stored
		err = txn(c, &modified, err == ds.ErrNoSuchEntity)
		if err != nil && err != errSkipPut {
			fatal = !errors.IsTransient(err)
			return err
		}
		if err != errSkipPut && !modified.isEqual(&stored) {
			return ds.Put(c, &modified)
		}
		return nil
	}, &defaultTransactionOptions)
	if err != nil {
		logging.Errorf(c, "Job transaction failed: %s", err)
		if fatal {
			return err
		}
		// By now err is already transient (since 'fatal' is false) or it is commit
		// error (i.e. produced by RunInTransaction itself, not by its callback).
		// Need to wrap commit errors too.
		return errors.WrapTransient(err)
	}
	if attempt > 1 {
		logging.Infof(c, "Committed on %d attempt", attempt)
	}
	return nil
}

// rollSM is called under transaction to perform a single state machine
// transition. It sets up StateMachine instance, calls the callback, mutates
// job.State in place (with a new state) and enqueues all emitted actions to
// task queues.
func (e *engineImpl) rollSM(c context.Context, job *Job, cb func(*StateMachine) error) error {
	sched, err := job.parseSchedule()
	if err != nil {
		return fmt.Errorf("bad schedule %q - %s", job.effectiveSchedule(), err)
	}
	now := clock.Now(c).UTC()
	rnd := mathrand.Get(c)
	sm := StateMachine{
		State:    job.State,
		Now:      now,
		Schedule: sched,
		Nonce:    func() int64 { return rnd.Int63() + 1 },
		Context:  c,
	}
	// All errors returned by state machine transition changes are transient.
	// Fatal errors (when we have them) should be reflected as a state changing
	// into "BROKEN" state.
	if err := cb(&sm); err != nil {
		return errors.WrapTransient(err)
	}
	if len(sm.Actions) != 0 {
		if err := e.enqueueJobActions(c, job.JobID, sm.Actions); err != nil {
			return err
		}
	}
	if sm.State.State != job.State.State {
		logging.Infof(c, "%s -> %s", job.State.State, sm.State.State)
	}
	job.State = sm.State
	return nil
}

// enqueueJobActions submits all actions emitted by a job state transition by
// adding corresponding tasks to task queues. See ExecuteSerializedAction for
// place where these actions are interpreted.
func (e *engineImpl) enqueueJobActions(c context.Context, jobID string, actions []Action) error {
	// AddMulti can't put tasks into multiple queues at once, split by queue name.
	qs := map[string][]*tq.Task{}
	for _, a := range actions {
		switch a := a.(type) {
		case TickLaterAction:
			payload, err := json.Marshal(actionTaskPayload{
				JobID:     jobID,
				Kind:      "TickLaterAction",
				TickNonce: a.TickNonce,
			})
			if err != nil {
				return err
			}
			logging.Infof(c, "Scheduling tick %d after %.1f sec", a.TickNonce, a.When.Sub(time.Now()).Seconds())
			qs[e.TimersQueueName] = append(qs[e.TimersQueueName], &tq.Task{
				Path:    e.TimersQueuePath,
				ETA:     a.When,
				Payload: payload,
			})
		case StartInvocationAction:
			payload, err := json.Marshal(actionTaskPayload{
				JobID:           jobID,
				Kind:            "StartInvocationAction",
				InvocationNonce: a.InvocationNonce,
				TriggeredBy:     string(a.TriggeredBy),
			})
			if err != nil {
				return err
			}
			qs[e.InvocationsQueueName] = append(qs[e.InvocationsQueueName], &tq.Task{
				Path:    e.InvocationsQueuePath,
				Delay:   time.Second, // give the transaction time to land
				Payload: payload,
				RetryOptions: &tq.RetryOptions{
					// Give 5 attempts to mark the job as failed. See 'startInvocation'.
					RetryLimit: invocationRetryLimit + 5,
					MinBackoff: time.Second,
					MaxBackoff: maxInvocationRetryBackoff,
					AgeLimit:   time.Duration(invocationRetryLimit+5) * maxInvocationRetryBackoff,
				},
			})
		case RecordOverrunAction:
			payload, err := json.Marshal(actionTaskPayload{
				JobID:               jobID,
				Kind:                "RecordOverrunAction",
				Overruns:            a.Overruns,
				RunningInvocationID: a.RunningInvocationID,
			})
			if err != nil {
				return err
			}
			qs[e.InvocationsQueueName] = append(qs[e.InvocationsQueueName], &tq.Task{
				Path:    e.InvocationsQueuePath,
				Delay:   time.Second, // give the transaction time to land
				Payload: payload,
			})
		default:
			logging.Errorf(c, "Unexpected action type %T, skipping", a)
		}
	}
	wg := sync.WaitGroup{}
	errs := errors.NewLazyMultiError(len(qs))
	i := 0
	for queueName, tasks := range qs {
		wg.Add(1)
		go func(i int, queueName string, tasks []*tq.Task) {
			errs.Assign(i, tq.Add(c, queueName, tasks...))
			wg.Done()
		}(i, queueName, tasks)
		i++
	}
	wg.Wait()
	return errors.WrapTransient(errs.Get())
}

// enqueueInvTimers submits all timers emitted by an invocation manager by
// adding corresponding tasks to the task queue. See ExecuteSerializedAction for
// place where these actions are interpreted.
func (e *engineImpl) enqueueInvTimers(c context.Context, inv *Invocation, timers []invocationTimer) error {
	tasks := make([]*tq.Task, len(timers))
	for i, timer := range timers {
		payload, err := json.Marshal(actionTaskPayload{
			JobID:    inv.JobKey.StringID(),
			InvID:    inv.ID,
			InvTimer: &timer,
		})
		if err != nil {
			return err
		}
		tasks[i] = &tq.Task{
			Path:    e.TimersQueuePath,
			ETA:     clock.Now(c).Add(timer.Delay),
			Payload: payload,
		}
	}
	return errors.WrapTransient(tq.Add(c, e.TimersQueueName, tasks...))
}

func (e *engineImpl) ExecuteSerializedAction(c context.Context, action []byte, retryCount int) error {
	payload := actionTaskPayload{}
	if err := json.Unmarshal(action, &payload); err != nil {
		return err
	}
	if payload.InvID == 0 {
		return e.executeJobAction(c, &payload, retryCount)
	}
	return e.executeInvAction(c, &payload, retryCount)
}

func (e *engineImpl) executeJobAction(c context.Context, payload *actionTaskPayload, retryCount int) error {
	switch payload.Kind {
	case "TickLaterAction":
		return e.jobTimerTick(c, payload.JobID, payload.TickNonce)
	case "StartInvocationAction":
		return e.startInvocation(
			c, payload.JobID, payload.InvocationNonce,
			identity.Identity(payload.TriggeredBy), retryCount)
	case "RecordOverrunAction":
		return e.recordOverrun(c, payload.JobID, payload.Overruns, payload.RunningInvocationID)
	default:
		return fmt.Errorf("unexpected job action kind %q", payload.Kind)
	}
}

func (e *engineImpl) executeInvAction(c context.Context, payload *actionTaskPayload, retryCount int) error {
	switch {
	case payload.InvTimer != nil:
		return e.invocationTimerTick(c, payload.JobID, payload.InvID, payload.InvTimer)
	default:
		return fmt.Errorf("unexpected invocation action kind %q", payload)
	}
}

func (e *engineImpl) TriggerInvocation(c context.Context, jobID string, triggeredBy identity.Identity) (int64, error) {
	var err error
	var invNonce int64
	err2 := e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew {
			err = errors.New("no such job")
			return errSkipPut
		}
		if !job.Enabled {
			err = errors.New("the job is disabled")
			return errSkipPut
		}
		invNonce = 0
		return e.rollSM(c, job, func(sm *StateMachine) error {
			if err := sm.OnManualInvocation(triggeredBy); err != nil {
				return err
			}
			invNonce = sm.State.InvocationNonce
			return nil
		})
	})
	if err == nil {
		err = err2
	}
	return invNonce, err
}

func (e *engineImpl) PauseJob(c context.Context, jobID string, who identity.Identity) error {
	return e.setPausedFlag(c, jobID, true, who)
}

func (e *engineImpl) ResumeJob(c context.Context, jobID string, who identity.Identity) error {
	return e.setPausedFlag(c, jobID, false, who)
}

func (e *engineImpl) setPausedFlag(c context.Context, jobID string, paused bool, who identity.Identity) error {
	return e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew || !job.Enabled {
			return errors.New("no such job")
		}
		if job.Paused == paused {
			return errSkipPut
		}
		if paused {
			logging.Warningf(c, "Job is paused by %s", who)
		} else {
			logging.Warningf(c, "Job is resumed by %s", who)
		}
		job.Paused = paused
		return e.rollSM(c, job, func(sm *StateMachine) error {
			sm.OnScheduleChange()
			return nil
		})
	})
}

func (e *engineImpl) AbortInvocation(c context.Context, jobID string, invID int64, who identity.Identity) error {
	c = logging.SetField(c, "JobID", jobID)
	c = logging.SetField(c, "InvID", invID)

	var inv *Invocation
	var err error
	switch inv, err = e.GetInvocation(c, jobID, invID); {
	case err != nil:
		logging.Errorf(c, "Failed to fetch the invocation - %s", err)
		return err
	case inv == nil:
		logging.Errorf(c, "The invocation doesn't exist")
		return errors.New("the invocation doesn't exist")
	case inv.Status.Final():
		return nil
	}

	ctl, err := e.controllerForInvocation(c, inv)
	if err != nil {
		logging.Errorf(c, "Cannot get controller - %s", err)
		return err
	}

	ctl.DebugLog("Invocation is manually aborted by %q", who)
	if err = ctl.manager.AbortTask(c, ctl); err != nil {
		logging.Errorf(c, "Failed to abort the task - %s", err)
		return err
	}

	ctl.State().Status = task.StatusAborted
	if err = ctl.Save(c); err != nil {
		logging.Errorf(c, "Failed to save the invocation - %s", err)
		return err
	}
	return nil
}

// AbortJob resets the job to scheduled state, aborting a currently pending or
// running invocation (if any).
//
// Returns nil if the job is not currently running.
func (e *engineImpl) AbortJob(c context.Context, jobID string, who identity.Identity) error {
	// First we switch the job to the default state and disassociate the running
	// invocation (if any) from the job entity.
	var invID int64
	err := e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew {
			return errSkipPut // the job was removed, nothing to abort
		}
		invID = job.State.InvocationID
		return e.rollSM(c, job, func(sm *StateMachine) error {
			sm.OnManualAbort()
			return nil
		})
	})
	if err != nil {
		return err
	}

	// Now we kill the invocation. We do it separately because it may involve
	// an RPC to remote service (e.g. to cancel a task) that can't be done from
	// the transaction.
	if invID != 0 {
		return e.AbortInvocation(c, jobID, invID, who)
	}
	return nil
}

// updateJob updates an existing job if its definition has changed, adds
// a completely new job or enables a previously disabled job.
func (e *engineImpl) updateJob(c context.Context, def catalog.Definition) error {
	return e.txn(c, def.JobID, func(c context.Context, job *Job, isNew bool) error {
		if !isNew && job.Enabled && job.matches(def) {
			return errSkipPut
		}
		if isNew {
			// JobID is <projectID>/<name>, it's ensure by Catalog.
			chunks := strings.Split(def.JobID, "/")
			if len(chunks) != 2 {
				return fmt.Errorf("unexpected jobID format: %s", def.JobID)
			}
			*job = Job{
				JobID:     def.JobID,
				ProjectID: chunks[0],
				Flavor:    def.Flavor,
				Enabled:   false, // to trigger 'if !oldEnabled' below
				Schedule:  def.Schedule,
				Task:      def.Task,
				State:     JobState{State: JobStateDisabled},
			}
		}
		oldEnabled := job.Enabled
		oldEffectiveSchedule := job.effectiveSchedule()

		// Update the job in full before running any state changes.
		job.Flavor = def.Flavor
		job.Revision = def.Revision
		job.RevisionURL = def.RevisionURL
		job.Enabled = true
		job.Schedule = def.Schedule
		job.Task = def.Task

		// Do state machine transitions.
		if !oldEnabled {
			err := e.rollSM(c, job, func(sm *StateMachine) error {
				sm.OnJobEnabled()
				return nil
			})
			if err != nil {
				return err
			}
		}
		if job.effectiveSchedule() != oldEffectiveSchedule {
			logging.Infof(c, "Job's schedule changed")
			return e.rollSM(c, job, func(sm *StateMachine) error {
				sm.OnScheduleChange()
				return nil
			})
		}
		return nil
	})
}

// disableJob moves a job to disabled state.
func (e *engineImpl) disableJob(c context.Context, jobID string) error {
	return e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew || !job.Enabled {
			return errSkipPut
		}
		job.Enabled = false
		return e.rollSM(c, job, func(sm *StateMachine) error {
			sm.OnJobDisabled()
			return nil
		})
	})
}

// resetJobOnDevServer sends "off" signal followed by "on" signal.
//
// It effectively cancels any pending actions and schedules new ones. Used only
// on dev server.
func (e *engineImpl) resetJobOnDevServer(c context.Context, jobID string) error {
	return e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew || !job.Enabled {
			return errSkipPut
		}
		logging.Infof(c, "Resetting job")
		err := e.rollSM(c, job, func(sm *StateMachine) error {
			sm.OnJobDisabled()
			return nil
		})
		if err != nil {
			return err
		}
		return e.rollSM(c, job, func(sm *StateMachine) error {
			sm.OnJobEnabled()
			return nil
		})
	})
}

// jobTimerTick is invoked via task queue in a task with some ETA. It what makes
// cron tick.
func (e *engineImpl) jobTimerTick(c context.Context, jobID string, tickNonce int64) error {
	return e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew {
			logging.Errorf(c, "Scheduled job is unexpectedly gone")
			return errSkipPut
		}
		logging.Infof(c, "Tick %d has arrived", tickNonce)
		return e.rollSM(c, job, func(sm *StateMachine) error { return sm.OnTimerTick(tickNonce) })
	})
}

// recordOverrun is invoked via task queue when a job should have been started,
// but previous invocation was still running.
//
// It creates new Invocation entity (in 'FAILED' state) in the datastore,
// to keep record of all overruns. Doesn't modify Job entity.
func (e *engineImpl) recordOverrun(c context.Context, jobID string, overruns int, runningInvID int64) error {
	now := clock.Now(c).UTC()
	jobKey := ds.NewKey(c, "Job", jobID, 0, nil)
	invID, err := generateInvocationID(c, jobKey)
	if err != nil {
		return err
	}
	inv := Invocation{
		ID:       invID,
		JobKey:   jobKey,
		Started:  now,
		Finished: now,
		Status:   task.StatusOverrun,
	}
	if runningInvID == 0 {
		inv.debugLog(c, "New invocation should be starting now, but previous one is still starting")
	} else {
		inv.debugLog(c, "New invocation should be starting now, but previous one is still running: %d", runningInvID)
	}
	inv.debugLog(c, "Total overruns thus far: %d", overruns)
	return errors.WrapTransient(ds.Put(c, &inv))
}

// invocationTimerTick is called via Task Queue to handle AddTimer callbacks.
//
// See also handlePubSubMessage, it is quite similar.
func (e *engineImpl) invocationTimerTick(c context.Context, jobID string, invID int64, timer *invocationTimer) error {
	c = logging.SetField(c, "JobID", jobID)
	c = logging.SetField(c, "InvID", invID)

	logging.Infof(c, "Handling invocation timer %q", timer.Name)
	inv, err := e.GetInvocation(c, jobID, invID)
	if err != nil {
		logging.Errorf(c, "Failed to fetch the invocation - %s", err)
		return err
	}
	if inv == nil {
		return errors.New("the invocation doesn't exist")
	}

	// Finished invocations are immutable, skip the message.
	if inv.Status.Final() {
		logging.Infof(c, "Skipping the timer, the invocation is in final state %q", inv.Status)
		return nil
	}

	// Build corresponding controller.
	ctl, err := e.controllerForInvocation(c, inv)
	if err != nil {
		logging.Errorf(c, "Cannot get controller - %s", err)
		return err
	}

	// Hand the message to the TaskManager.
	err = ctl.manager.HandleTimer(c, ctl, timer.Name, timer.Payload)
	if err != nil {
		logging.Errorf(c, "Error when handling the timer - %s", err)
		if !errors.IsTransient(err) && ctl.State().Status != task.StatusFailed {
			ctl.DebugLog("Fatal error when handling timer, aborting invocation - %s", err)
			ctl.State().Status = task.StatusFailed
		}
	}

	// Save anyway, to preserve the invocation log.
	saveErr := ctl.Save(c)
	if saveErr != nil {
		logging.Errorf(c, "Error when saving the invocation - %s", saveErr)
	}

	// Retry the delivery if at least one error is transient. HandleTimer must be
	// idempotent.
	switch {
	case err == nil && saveErr == nil:
		return nil
	case errors.IsTransient(saveErr):
		return saveErr
	default:
		return err // transient or fatal
	}
}

// startInvocation is called via task queue to start running a job. This call
// may be retried by task queue service.
func (e *engineImpl) startInvocation(c context.Context, jobID string, invocationNonce int64,
	triggeredBy identity.Identity, retryCount int) error {

	c = logging.SetField(c, "JobID", jobID)
	c = logging.SetField(c, "InvNonce", invocationNonce)
	c = logging.SetField(c, "Attempt", retryCount)

	// Create new Invocation entity in StatusStarting state and associated it with
	// Job entity.
	//
	// Task queue guarantees not to execute same task concurrently (i.e. retry
	// happens only if previous attempt finished already).
	// There are 4 possibilities here:
	// 1) It is a first attempt. In that case we generate new Invocation in
	//    state STARTING and update Job with a reference to it.
	// 2) It is a retry and previous attempt is still starting (indicated by
	//    IsExpectingInvocation returning true). Assume it failed to start
	//    and launch a new one. Mark old one as obsolete.
	// 3) It is a retry and previous attempt has already started (in this case
	//    the job is in RUNNING state and IsExpectingInvocation returns
	//    false). Assume this retry was unnecessary and skip it.
	// 4) It is a retry and we have retried too many times. Mark the invocation
	//    as failed and resume the job's schedule. This branch executes only if
	//    retryCount check at the end of this function failed to run (e.g. request
	//    handler crashed).
	inv := Invocation{}
	skipRunning := false
	err := e.txn(c, jobID, func(c context.Context, job *Job, isNew bool) error {
		if isNew {
			logging.Errorf(c, "Queued job is unexpectedly gone")
			skipRunning = true
			return errSkipPut
		}
		if !job.State.IsExpectingInvocation(invocationNonce) {
			logging.Errorf(c, "No longer need to start invocation with nonce %d", invocationNonce)
			skipRunning = true
			return errSkipPut
		}
		jobKey := ds.KeyForObj(c, job)
		invID, err := generateInvocationID(c, jobKey)
		if err != nil {
			return err
		}
		// Put new invocation entity, generate its ID.
		inv = Invocation{
			ID:              invID,
			JobKey:          jobKey,
			Started:         clock.Now(c).UTC(),
			InvocationNonce: invocationNonce,
			TriggeredBy:     triggeredBy,
			Revision:        job.Revision,
			RevisionURL:     job.RevisionURL,
			Task:            job.Task,
			RetryCount:      int64(retryCount),
			Status:          task.StatusStarting,
		}
		inv.debugLog(c, "Invocation initiated (attempt %d)", retryCount+1)
		if triggeredBy != "" {
			inv.debugLog(c, "Manually triggered by %s", triggeredBy)
		}
		if retryCount >= invocationRetryLimit {
			logging.Errorf(c, "Too many attempts, giving up")
			inv.debugLog(c, "Too many attempts, giving up")
			inv.Status = task.StatusFailed
			inv.Finished = clock.Now(c).UTC()
			skipRunning = true
		}
		if err := ds.Put(c, &inv); err != nil {
			return err
		}
		// Move previous invocation (if any) to failed state. It has failed to
		// start.
		if job.State.InvocationID != 0 {
			prev := Invocation{
				ID:     job.State.InvocationID,
				JobKey: jobKey,
			}
			err := ds.Get(c, &prev)
			if err != nil && err != ds.ErrNoSuchEntity {
				return err
			}
			if err == nil && !prev.Status.Final() {
				prev.debugLog(c, "New invocation is starting (%d), marking this one as failed.", inv.ID)
				prev.Status = task.StatusFailed
				prev.Finished = clock.Now(c).UTC()
				prev.MutationsCount++
				prev.trimDebugLog()
				if err := ds.Put(c, &prev); err != nil {
					return err
				}
			}
		}
		// Store the reference to the new invocation ID. Unblock the job if we are
		// giving up on retrying.
		return e.rollSM(c, job, func(sm *StateMachine) error {
			sm.OnInvocationStarting(invocationNonce, inv.ID, retryCount)
			if inv.Status == task.StatusFailed {
				sm.OnInvocationStarted(inv.ID)
				sm.OnInvocationDone(inv.ID)
			}
			return nil
		})
	})

	if err != nil {
		logging.WithError(err).Errorf(c, "Failed to update job state")
		return err
	}

	if skipRunning {
		logging.Warningf(c, "No need to start the invocation anymore")
		return nil
	}

	c = logging.SetField(c, "InvID", inv.ID)

	// Now we have a new Invocation entity in the datastore in StatusStarting
	// state. Grab corresponding TaskManager and launch task through it, keeping
	// track of the progress in created Invocation entity.
	ctl, err := e.controllerForInvocation(c, &inv)
	if err != nil {
		// Note: controllerForInvocation returns both ctl and err on errors, with
		// ctl not fully initialized (but good enough for what's done below).
		ctl.DebugLog("Failed to initialize task controller - %s", err)
		ctl.State().Status = task.StatusFailed
		return ctl.Save(c)
	}

	// Ask manager to start the task. If it returns no errors, it should also move
	// invocation out of StatusStarting state (a failure to do so is an error). If
	// it returns an error, invocation is forcefully moved to StatusFailed state.
	// In either case, invocation never ends up in StatusStarting state.
	err = ctl.manager.LaunchTask(c, ctl)
	if ctl.State().Status == task.StatusStarting {
		ctl.State().Status = task.StatusFailed
		if err == nil {
			err = fmt.Errorf("LaunchTask didn't move invocation out of StatusStarting")
		}
	}

	// Give up retrying on transient errors after some number of attempts.
	if errors.IsTransient(err) && retryCount+1 >= invocationRetryLimit {
		err = fmt.Errorf("Too many retries, giving up (original error - %s)", err)
	}

	// If asked to retry the invocation (by returning a transient error), do not
	// touch Job entity when saving the current (failed) invocation. That way Job
	// stays in "QUEUED" state (indicating it's queued for a new invocation).
	retryInvocation := errors.IsTransient(err)
	if saveErr := ctl.saveImpl(c, !retryInvocation); saveErr != nil {
		logging.Errorf(c, "Failed to save invocation state - %s", saveErr)
		if err == nil {
			err = saveErr
		}
	}

	// Returning transient error here causes the task queue to retry the task.
	if err != nil {
		logging.WithError(err).Errorf(c, "Invocation failed to start")
	}
	return err
}

// controllerForInvocation returns new instance of taskController configured
// to work with given invocation.
//
// If task definition can't be deserialized, returns both controller and error.
func (e *engineImpl) controllerForInvocation(c context.Context, inv *Invocation) (*taskController, error) {
	ctl := &taskController{
		ctx:   c, // for DebugLog
		eng:   e,
		saved: *inv,
	}
	ctl.populateState()
	var err error
	ctl.task, err = e.Catalog.UnmarshalTask(inv.Task)
	if err != nil {
		return ctl, fmt.Errorf("failed to unmarshal the task - %s", err)
	}
	ctl.manager = e.Catalog.GetTaskManager(ctl.task)
	if ctl.manager == nil {
		return ctl, fmt.Errorf("TaskManager is unexpectedly missing")
	}
	return ctl, nil
}

////////////////////////////////////////////////////////////////////////////////
// PubSub stuff

// topicParams is passed to prepareTopic by task.Controller.
type topicParams struct {
	inv       *Invocation  // invocation being handled by Controller
	manager   task.Manager // task manager for the invocation
	publisher string       // name of publisher to add to PubSub topic.
}

// pubsubAuthToken describes how to generate HMAC protected tokens used to
// authenticate PubSub messages.
var pubsubAuthToken = tokens.TokenKind{
	Algo:       tokens.TokenAlgoHmacSHA256,
	Expiration: 48 * time.Hour,
	SecretKey:  "pubsub_auth_token",
	Version:    1,
}

// genTopicAndSubNames derives PubSub topic and subscription names to use for
// notifications from given publisher.
func (e *engineImpl) genTopicAndSubNames(c context.Context, manager, publisher string) (topic string, sub string) {
	// Avoid accidental override of the topic when running on dev server.
	prefix := "scheduler"
	if info.IsDevAppServer(c) {
		prefix = "dev-scheduler"
	}

	// Each publisher gets its own topic (and subscription), so it's clearer from
	// logs and PubSub console who's calling what. PubSub topics can't have "@" in
	// them, so replace "@" with "~". URL encoding could have been used too, but
	// Cloud Console confuses %40 with its own URL encoding and doesn't display
	// all pages correctly.
	id := fmt.Sprintf("%s+%s+%s",
		prefix,
		manager,
		strings.Replace(publisher, "@", "~", -1))

	appID := info.AppID(c)
	topic = fmt.Sprintf("projects/%s/topics/%s", appID, id)
	sub = fmt.Sprintf("projects/%s/subscriptions/%s", appID, id)
	return
}

// prepareTopic creates a pubsub topic that can be used to pass task related
// messages back to the task.Manager that handles the task.
//
// It returns full topic name, as well as a token that securely identifies the
// task. It should be put into 'auth_token' attribute of PubSub messages by
// whoever publishes them.
func (e *engineImpl) prepareTopic(c context.Context, params topicParams) (topic string, tok string, err error) {
	// If given URL, ask the service for name of its default service account.
	// FetchServiceInfo implements efficient cache internally, so it's fine to
	// call it often.
	if strings.HasPrefix(params.publisher, "https://") {
		logging.Infof(c, "Fetching info about %q", params.publisher)
		serviceInfo, err := signing.FetchServiceInfoFromLUCIService(c, params.publisher)
		if err != nil {
			logging.Errorf(c, "Failed to fetch info about %q - %s", params.publisher, err)
			return "", "", err
		}
		logging.Infof(c, "%q is using %q", params.publisher, serviceInfo.ServiceAccountName)
		params.publisher = serviceInfo.ServiceAccountName
	}

	topic, sub := e.genTopicAndSubNames(c, params.manager.Name(), params.publisher)

	// Put same parameters in push URL to make them visible in logs. On dev server
	// use pull based subscription, since localhost push URL is not valid.
	pushURL := ""
	if !info.IsDevAppServer(c) {
		urlParams := url.Values{}
		urlParams.Add("kind", params.manager.Name())
		urlParams.Add("publisher", params.publisher)
		pushURL = fmt.Sprintf(
			"https://%s%s?%s", info.DefaultVersionHostname(c), e.PubSubPushPath, urlParams.Encode())
	}

	// Create and configure the topic. Do it only once.
	err = e.doIfNotDone(c, fmt.Sprintf("prepareTopic:v1:%s", topic), func() error {
		if e.configureTopic != nil {
			return e.configureTopic(c, topic, sub, pushURL, params.publisher)
		}
		return configureTopic(c, topic, sub, pushURL, params.publisher, "")
	})
	if err != nil {
		return "", "", err
	}

	// Encode full invocation identifier (job key + invocation ID) into HMAC
	// protected token.
	tok, err = pubsubAuthToken.Generate(c, nil, map[string]string{
		"job": params.inv.JobKey.StringID(),
		"inv": fmt.Sprintf("%d", params.inv.ID),
	}, 0)
	if err != nil {
		return "", "", err
	}

	return topic, tok, nil
}

func (e *engineImpl) ProcessPubSubPush(c context.Context, body []byte) error {
	var pushBody struct {
		Message pubsub.PubsubMessage `json:"message"`
	}
	if err := json.Unmarshal(body, &pushBody); err != nil {
		return err
	}
	return e.handlePubSubMessage(c, &pushBody.Message)
}

func (e *engineImpl) PullPubSubOnDevServer(c context.Context, taskManagerName, publisher string) error {
	_, sub := e.genTopicAndSubNames(c, taskManagerName, publisher)
	msg, ack, err := pullSubcription(c, sub, "")
	if err != nil {
		return err
	}
	if msg == nil {
		logging.Infof(c, "No new PubSub messages")
		return nil
	}
	err = e.handlePubSubMessage(c, msg)
	if err == nil || !errors.IsTransient(err) {
		ack() // ack only on success of fatal errors (to stop redelivery)
	}
	return err
}

func (e *engineImpl) handlePubSubMessage(c context.Context, msg *pubsub.PubsubMessage) error {
	logging.Infof(c, "Received PubSub message %q", msg.MessageId)

	// Extract Job and Invocation ID from validated auth_token.
	var jobID string
	var invID int64
	data, err := pubsubAuthToken.Validate(c, msg.Attributes["auth_token"], nil)
	if err != nil {
		logging.Errorf(c, "Bad auth_token attribute - %s", err)
		return err
	}
	jobID = data["job"]
	if invID, err = strconv.ParseInt(data["inv"], 10, 64); err != nil {
		logging.Errorf(c, "Could not parse 'inv' %q - %s", data["inv"], err)
		return err
	}

	c = logging.SetField(c, "JobID", jobID)
	c = logging.SetField(c, "InvID", invID)
	inv, err := e.GetInvocation(c, jobID, invID)
	if err != nil {
		logging.Errorf(c, "Failed to fetch the invocation - %s", err)
		return err
	}
	if inv == nil {
		return errors.New("the invocation doesn't exist")
	}

	// Finished invocations are immutable, skip the message.
	if inv.Status.Final() {
		logging.Infof(c, "Skipping the notification, the invocation is in final state %q", inv.Status)
		return nil
	}

	// Build corresponding controller.
	ctl, err := e.controllerForInvocation(c, inv)
	if err != nil {
		logging.Errorf(c, "Cannot get controller - %s", err)
		return err
	}

	// Hand the message to the TaskManager.
	err = ctl.manager.HandleNotification(c, ctl, msg)
	if err != nil {
		logging.Errorf(c, "Error when handling the message - %s", err)
		if !errors.IsTransient(err) && ctl.State().Status != task.StatusFailed {
			ctl.DebugLog("Fatal error when handling PubSub notification, aborting invocation - %s", err)
			ctl.State().Status = task.StatusFailed
		}
	}

	// Save anyway, to preserve the invocation log.
	saveErr := ctl.Save(c)
	if saveErr != nil {
		logging.Errorf(c, "Error when saving the invocation - %s", saveErr)
	}

	// Retry the delivery if at least one error is transient. HandleNotification
	// must be idempotent.
	switch {
	case err == nil && saveErr == nil:
		return nil
	case errors.IsTransient(saveErr):
		return saveErr
	default:
		return err // transient or fatal
	}
}

////////////////////////////////////////////////////////////////////////////////
// TaskController.

type taskController struct {
	ctx     context.Context
	eng     *engineImpl
	manager task.Manager
	task    proto.Message // extracted from saved.Task blob

	saved    Invocation        // what have been given initially or saved in Save()
	state    task.State        // state mutated by TaskManager
	debugLog string            // mutated by DebugLog
	timers   []invocationTimer // mutated by AddTimer
}

// populateState populates 'state' using data in 'saved'.
func (ctl *taskController) populateState() {
	ctl.state = task.State{
		Status:   ctl.saved.Status,
		ViewURL:  ctl.saved.ViewURL,
		TaskData: append([]byte(nil), ctl.saved.TaskData...), // copy
	}
}

// JobID is part of task.Controller interface.
func (ctl *taskController) JobID() string {
	return ctl.saved.JobKey.StringID()
}

// InvocationID is part of task.Controller interface.
func (ctl *taskController) InvocationID() int64 {
	return ctl.saved.ID
}

// InvocationNonce is part of task.Controller interface.
func (ctl *taskController) InvocationNonce() int64 {
	return ctl.saved.InvocationNonce
}

// Task is part of task.Controller interface.
func (ctl *taskController) Task() proto.Message {
	return ctl.task
}

// State is part of task.Controller interface.
func (ctl *taskController) State() *task.State {
	return &ctl.state
}

// DebugLog is part of task.Controller interface.
func (ctl *taskController) DebugLog(format string, args ...interface{}) {
	logging.Infof(ctl.ctx, format, args...)
	debugLog(ctl.ctx, &ctl.debugLog, format, args...)
}

// AddTimer is part of Controller interface.
func (ctl *taskController) AddTimer(ctx context.Context, delay time.Duration, name string, payload []byte) {
	ctl.DebugLog("Scheduling timer %q after %s", name, delay)
	ctl.timers = append(ctl.timers, invocationTimer{
		Delay:   delay,
		Name:    name,
		Payload: payload,
	})
}

// PrepareTopic is part of task.Controller interface.
func (ctl *taskController) PrepareTopic(ctx context.Context, publisher string) (topic string, token string, err error) {
	return ctl.eng.prepareTopic(ctx, topicParams{
		inv:       &ctl.saved,
		manager:   ctl.manager,
		publisher: publisher,
	})
}

// GetClient is part of task.Controller interface
func (ctl *taskController) GetClient(ctx context.Context, timeout time.Duration) (*http.Client, error) {
	// TODO(vadimsh): Use per-project service accounts, not a global service
	// account.
	ctx, _ = clock.WithTimeout(ctx, timeout)
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: t}, nil
}

// Save is part of task.Controller interface.
func (ctl *taskController) Save(ctx context.Context) error {
	return ctl.saveImpl(ctx, true)
}

// errUpdateConflict means Invocation is being modified by two TaskController's
// concurrently. It should not be happening often. If it happens, task queue
// call is retried to rerun the two-part transaction from scratch.
var errUpdateConflict = errors.WrapTransient(errors.New("concurrent modifications of single Invocation"))

// saveImpl uploads updated Invocation to the datastore. If updateJob is true,
// it will also roll corresponding state machine forward.
func (ctl *taskController) saveImpl(ctx context.Context, updateJob bool) (err error) {
	// Mutate copy in case transaction below fails. Also unpacks ctl.state back
	// into the entity (reverse of 'populateState').
	saving := ctl.saved
	saving.Status = ctl.state.Status
	saving.TaskData = append([]byte(nil), ctl.state.TaskData...)
	saving.ViewURL = ctl.state.ViewURL
	saving.DebugLog += ctl.debugLog
	if saving.isEqual(&ctl.saved) && len(ctl.timers) == 0 { // no changes at all?
		return nil
	}
	saving.MutationsCount++

	// Update local copy of Invocation with what's in the datastore on success.
	defer func() {
		if err == nil {
			ctl.saved = saving
			ctl.debugLog = "" // debug log was successfully flushed
			ctl.timers = nil  // timers were successfully scheduled
		}
	}()

	hasStartedOrFailed := ctl.saved.Status == task.StatusStarting && saving.Status != task.StatusStarting
	hasFinished := !ctl.saved.Status.Final() && saving.Status.Final()
	if hasFinished {
		saving.Finished = clock.Now(ctx)
		saving.debugLog(
			ctx, "Invocation finished in %s with status %s",
			saving.Finished.Sub(saving.Started), saving.Status)
		if !updateJob {
			saving.debugLog(ctx, "It will probably be retried")
		}
		// Finished invocations can't schedule timers.
		for _, t := range ctl.timers {
			saving.debugLog(ctx, "Ignoring timer %s...", t.Name)
		}
	}

	// Store the invocation entity, mutate Job state accordingly, schedule all
	// timer ticks.
	return ctl.eng.txn(ctx, saving.JobKey.StringID(), func(c context.Context, job *Job, isNew bool) error {
		// Grab what's currently in the store to compare MutationsCount to what we
		// expect it to be.
		mostRecent := Invocation{
			ID:     saving.ID,
			JobKey: saving.JobKey,
		}
		switch err := ds.Get(c, &mostRecent); {
		case err == ds.ErrNoSuchEntity: // should not happen
			logging.Errorf(c, "Invocation is suddenly gone")
			return errors.New("invocation is suddenly gone")
		case err != nil:
			return errors.WrapTransient(err)
		}

		// Make sure no one touched it while we were handling the invocation.
		if saving.MutationsCount != mostRecent.MutationsCount+1 {
			logging.Errorf(c, "Invocation was modified by someone else while we were handling it")
			return errUpdateConflict
		}

		// Store the invocation entity and schedule invocation timers regardless of
		// the current state of the Job entity. The table of all invocations is
		// useful on its own (e.g. for debugging) even if Job entity state has
		// desynchronized for some reason.
		saving.trimDebugLog()
		if err := ds.Put(c, &saving); err != nil {
			return err
		}

		// Finished invocations can't schedule timers.
		if !hasFinished && len(ctl.timers) > 0 {
			if err := ctl.eng.enqueueInvTimers(c, &saving, ctl.timers); err != nil {
				return err
			}
		}

		// Is Job entity still have this invocation as a current one?
		switch {
		case !updateJob:
			logging.Warningf(c, "Asked not to touch Job entity")
			return errSkipPut
		case isNew:
			logging.Errorf(c, "Active job is unexpectedly gone")
			return errSkipPut
		case job.State.InvocationID != saving.ID:
			logging.Warningf(c, "The invocation is no longer current, the current is %d", job.State.InvocationID)
			return errSkipPut
		}

		// Make the state machine transitions.
		if hasStartedOrFailed {
			err := ctl.eng.rollSM(c, job, func(sm *StateMachine) error {
				sm.OnInvocationStarted(saving.ID)
				return nil
			})
			if err != nil {
				return err
			}
		}
		if hasFinished {
			return ctl.eng.rollSM(c, job, func(sm *StateMachine) error {
				sm.OnInvocationDone(saving.ID)
				return nil
			})
		}
		return nil
	})
}
