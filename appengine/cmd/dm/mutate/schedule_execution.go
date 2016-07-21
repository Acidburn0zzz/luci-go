// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mutate

import (
	"fmt"

	"github.com/luci/gae/filter/txnBuf"
	"github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/appengine/cmd/dm/distributor"
	"github.com/luci/luci-go/appengine/cmd/dm/model"
	"github.com/luci/luci-go/appengine/tumble"
	dm "github.com/luci/luci-go/common/api/dm/service/v1"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"
	"golang.org/x/net/context"
)

// ScheduleExecution is a placeholder mutation that will be an entry into the
// Distributor scheduling state-machine.
type ScheduleExecution struct {
	For *dm.Attempt_ID
}

// Root implements tumble.Mutation
func (s *ScheduleExecution) Root(c context.Context) *datastore.Key {
	return model.AttemptKeyFromID(c, s.For)
}

// RollForward implements tumble.Mutation
func (s *ScheduleExecution) RollForward(c context.Context) (muts []tumble.Mutation, err error) {
	ds := datastore.Get(c)
	a := model.AttemptFromID(s.For)
	if err = ds.Get(a); err != nil {
		return
	}

	if a.State != dm.Attempt_SCHEDULING {
		return
	}

	q := model.QuestFromID(s.For.Quest)
	if err = txnBuf.GetNoTxn(c).Get(q); err != nil {
		return
	}

	prevResult := (*dm.JsonResult)(nil)
	if a.LastSuccessfulExecution != 0 {
		prevExecution := model.ExecutionFromID(c, s.For.Execution(a.LastSuccessfulExecution))
		if err = ds.Get(prevExecution); err != nil {
			return
		}
		prevResult = prevExecution.Result.Data
	}

	reg := distributor.GetRegistry(c)
	dist, ver, err := reg.MakeDistributor(c, q.Desc.DistributorConfigName)
	if err != nil {
		return
	}

	a.CurExecution++
	if err = a.ModifyState(c, dm.Attempt_EXECUTING); err != nil {
		return
	}

	eid := dm.NewExecutionID(s.For.Quest, s.For.Id, a.CurExecution)
	e := model.MakeExecution(c, eid, q.Desc.DistributorConfigName, ver)

	exAuth := &dm.Execution_Auth{Id: eid, Token: e.Token}

	var distTok distributor.Token
	distTok, e.TimeToStart, e.TimeToRun, e.TimeToStop, err = dist.Run(
		distributor.NewTaskDescription(c, &q.Desc, exAuth, prevResult))
	e.DistributorToken = string(distTok)
	if err != nil {
		if errors.IsTransient(err) {
			// tumble will retry us later
			logging.WithError(err).Errorf(c, "got transient error in ScheduleExecution")
			return
		}
		origErr := err

		// put a and e to the transaction buffer, so that
		// FinishExecution.RollForward can see them.
		if err = ds.Put(a, e); err != nil {
			return
		}
		return NewFinishExecutionAbnormal(
			eid, dm.AbnormalFinish_REJECTED,
			fmt.Sprintf("rejected during scheduling with non-transient error: %s", origErr),
		).RollForward(c)
	}

	if err = ResetExecutionTimeout(c, e); err != nil {
		return
	}

	err = ds.Put(a, e)
	return
}

func init() {
	tumble.Register((*ScheduleExecution)(nil))
}
