// Copyright 2015 The LUCI Authors.
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

package errors

import (
	"errors"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLazyMultiError(t *testing.T) {
	t.Parallel()

	Convey("Test LazyMultiError", t, func() {
		lme := NewLazyMultiError(10)
		So(lme.Get(), ShouldEqual, nil)

		e := errors.New("sup")
		lme.Assign(6, e)
		So(lme.Get(), ShouldResemble,
			MultiError{nil, nil, nil, nil, nil, nil, e, nil, nil, nil})

		lme.Assign(2, e)
		So(lme.Get(), ShouldResemble,
			MultiError{nil, nil, e, nil, nil, nil, e, nil, nil, nil})

		So(func() { lme.Assign(20, e) }, ShouldPanic)

		Convey("Try to freak out the race detector", func() {
			lme := NewLazyMultiError(64)
			Convey("all nils", func() {
				wg := sync.WaitGroup{}
				for i := 0; i < 64; i++ {
					wg.Add(1)
					go func(i int) {
						lme.Assign(i, nil)
						wg.Done()
					}(i)
				}
				wg.Wait()
				So(lme.Get(), ShouldBeNil)
			})
			Convey("every other", func() {
				wow := errors.New("wow")
				wg := sync.WaitGroup{}
				for i := 0; i < 64; i++ {
					wg.Add(1)
					go func(i int) {
						e := error(nil)
						if i&1 == 1 {
							e = wow
						}
						lme.Assign(i, e)
						wg.Done()
					}(i)
				}
				wg.Wait()
				me := make(MultiError, 64)
				for i := range me {
					if i&1 == 1 {
						me[i] = wow
					}
				}
				So(lme.Get(), ShouldResemble, me)
			})
			Convey("all", func() {
				wow := errors.New("wow")
				wg := sync.WaitGroup{}
				for i := 0; i < 64; i++ {
					wg.Add(1)
					go func(i int) {
						lme.Assign(i, wow)
						wg.Done()
					}(i)
				}
				wg.Wait()
				me := make(MultiError, 64)
				for i := range me {
					me[i] = wow
				}
				So(lme.Get(), ShouldResemble, me)
			})
		})
	})
}
