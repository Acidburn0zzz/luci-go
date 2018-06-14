// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/grpc/grpcutil"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestTags(t *testing.T) {
	t.Parallel()

	Convey("With datastore", t, func() {
		digest := strings.Repeat("a", 40)

		testTime := testclock.TestRecentTimeUTC.Round(time.Millisecond)
		ctx, _ := testclock.UseTime(gaetesting.TestingContext(), testTime)

		putInst := func(pkg, iid string, pendingProcs []string) *Instance {
			inst := &Instance{
				InstanceID:        iid,
				Package:           PackageKey(ctx, pkg),
				ProcessorsPending: pendingProcs,
			}
			So(datastore.Put(ctx, &Package{Name: pkg}, inst), ShouldBeNil)
			return inst
		}

		getTag := func(inst *Instance, tag string) *Tag {
			t := &Tag{ID: TagID(tag), Instance: datastore.KeyForObj(ctx, inst)}
			err := datastore.Get(ctx, t)
			if err == datastore.ErrNoSuchEntity {
				return nil
			}
			So(err, ShouldBeNil)
			return t
		}

		expectedTag := func(inst *Instance, tag string, who string) *Tag {
			return &Tag{
				ID:           TagID(tag),
				Instance:     datastore.KeyForObj(ctx, inst),
				Tag:          tag,
				RegisteredBy: who,
				RegisteredTs: testTime,
			}
		}

		Convey("AttachTags happy paths", func() {
			inst := putInst("pkg", digest, nil)

			// Attach one tag and verify it exist.
			So(AttachTags(ctx, inst, []string{"a:0"}, "user:abc@example.com"), ShouldBeNil)
			So(getTag(inst, "a:0"), ShouldResemble, expectedTag(inst, "a:0", "user:abc@example.com"))

			// Attach few more at once.
			So(AttachTags(ctx, inst, []string{"a:1", "a:2"}, "user:abc@example.com"), ShouldBeNil)
			So(getTag(inst, "a:1"), ShouldResemble, expectedTag(inst, "a:1", "user:abc@example.com"))
			So(getTag(inst, "a:2"), ShouldResemble, expectedTag(inst, "a:2", "user:abc@example.com"))

			// Try to reattach an existing one (notice the change in the email),
			// should be ignored.
			So(AttachTags(ctx, inst, []string{"a:0"}, "user:def@example.com"), ShouldBeNil)
			So(getTag(inst, "a:0"), ShouldResemble, expectedTag(inst, "a:0", "user:abc@example.com"))

			// Try to reattach a bunch of existing ones at once.
			So(AttachTags(ctx, inst, []string{"a:1", "a:2"}, "user:def@example.com"), ShouldBeNil)
			So(getTag(inst, "a:1"), ShouldResemble, expectedTag(inst, "a:1", "user:abc@example.com"))
			So(getTag(inst, "a:2"), ShouldResemble, expectedTag(inst, "a:2", "user:abc@example.com"))

			// Mixed group with new and existing tags.
			So(AttachTags(ctx, inst, []string{"a:3", "a:0", "a:4", "a:1"}, "user:def@example.com"), ShouldBeNil)
			So(getTag(inst, "a:3"), ShouldResemble, expectedTag(inst, "a:3", "user:def@example.com"))
			So(getTag(inst, "a:0"), ShouldResemble, expectedTag(inst, "a:0", "user:abc@example.com"))
			So(getTag(inst, "a:4"), ShouldResemble, expectedTag(inst, "a:4", "user:def@example.com"))
			So(getTag(inst, "a:1"), ShouldResemble, expectedTag(inst, "a:1", "user:abc@example.com"))
		})

		Convey("DetachTags happy paths", func() {
			inst := putInst("pkg", digest, nil)

			// Attach a bunch of tags first, so we have something to detach.
			So(AttachTags(ctx, inst, []string{"a:0", "a:1", "a:2", "a:3", "a:4"}, "user:abc@example.com"), ShouldBeNil)

			// Detaching one existing.
			So(getTag(inst, "a:0"), ShouldNotBeNil)
			So(DetachTags(ctx, inst, []string{"a:0"}), ShouldBeNil)
			So(getTag(inst, "a:0"), ShouldBeNil)

			// Detaching one missing.
			So(DetachTags(ctx, inst, []string{"a:z0"}), ShouldBeNil)

			// Detaching a bunch of existing.
			So(getTag(inst, "a:1"), ShouldNotBeNil)
			So(getTag(inst, "a:2"), ShouldNotBeNil)
			So(DetachTags(ctx, inst, []string{"a:1", "a:2"}), ShouldBeNil)
			So(getTag(inst, "a:1"), ShouldBeNil)
			So(getTag(inst, "a:2"), ShouldBeNil)

			// Detaching a bunch of missing.
			So(DetachTags(ctx, inst, []string{"a:z1", "a:z2"}), ShouldBeNil)

			// Detaching a mix of existing and missing.
			So(getTag(inst, "a:3"), ShouldNotBeNil)
			So(getTag(inst, "a:4"), ShouldNotBeNil)
			So(DetachTags(ctx, inst, []string{"a:z3", "a:3", "a:z4", "a:4"}), ShouldBeNil)
			So(getTag(inst, "a:3"), ShouldBeNil)
			So(getTag(inst, "a:4"), ShouldBeNil)
		})

		Convey("AttachTags to not ready instance", func() {
			inst := putInst("pkg", digest, []string{"proc"})

			err := AttachTags(ctx, inst, []string{"a:0"}, "user:abc@example.com")
			So(grpcutil.Code(err), ShouldEqual, codes.FailedPrecondition)
			So(err, ShouldErrLike, "the instance is not ready yet")
		})

		Convey("Handles SHA1 collision", func() {
			inst := putInst("pkg", digest, nil)

			// We fake a collision here. Coming up with a real SHA1 collision to use
			// in this test is left as an exercise to the reader.
			So(datastore.Put(ctx, &Tag{
				ID:       TagID("some:tag"),
				Instance: datastore.KeyForObj(ctx, inst),
				Tag:      "another:tag",
			}), ShouldBeNil)

			Convey("AttachTags", func() {
				err := AttachTags(ctx, inst, []string{"some:tag"}, "user:abc@example.com")
				So(grpcutil.Code(err), ShouldEqual, codes.Internal)
				So(err, ShouldErrLike, `tag "some:tag" collides with tag "another:tag", refusing to touch it`)
			})

			Convey("DetachTags", func() {
				err := DetachTags(ctx, inst, []string{"some:tag"})
				So(grpcutil.Code(err), ShouldEqual, codes.Internal)
				So(err, ShouldErrLike, `tag "some:tag" collides with tag "another:tag", refusing to touch it`)
			})
		})
	})
}
