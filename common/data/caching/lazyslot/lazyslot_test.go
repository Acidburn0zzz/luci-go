// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package lazyslot

import (
	"errors"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/luci/luci-go/common/clock/testclock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLazySlot(t *testing.T) {
	Convey("Blocking mode works", t, func() {
		c, clk := newContext()

		lock := sync.Mutex{}
		counter := 0

		s := Slot{
			Fetcher: func(c context.Context, prev Value) (Value, error) {
				lock.Lock()
				defer lock.Unlock()
				counter++
				return Value{counter, clk.Now().Add(time.Second)}, nil
			},
		}

		// Initial fetch.
		So(s.current, ShouldBeNil)
		v, err := s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 1)

		// Still fresh.
		v, err = s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 1)

		// Expires and refreshed.
		clk.Add(5 * time.Second)
		v, err = s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 2)
	})

	Convey("Returns stale copy while fetching", t, func(conv C) {
		c, clk := newContext()

		// Put initial value.
		s := Slot{
			Fetcher: func(c context.Context, prev Value) (Value, error) {
				return Value{1, clk.Now().Add(time.Second)}, nil
			},
		}
		v, err := s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 1)

		// Make it expire. Start blocking fetch of the new value.
		clk.Add(5 * time.Second)
		fetching := make(chan bool)
		resume := make(chan bool)
		s.Fetcher = func(c context.Context, prev Value) (Value, error) {
			fetching <- true
			<-resume
			return Value{2, clk.Now().Add(time.Second)}, nil
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := s.Get(c)
			conv.So(err, ShouldBeNil)
			conv.So(v.Value.(int), ShouldEqual, 2)
		}()

		// Wait until we hit the body of the fetcher callback.
		<-fetching

		// Concurrent Get() returns stale copy right away (does not deadlock).
		v, err = s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 1)

		// Wait until another goroutine finishes the fetch.
		resume <- true
		wg.Wait()

		// Returns new value now.
		v, err = s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 2)
	})

	Convey("Recovers from panic", t, func() {
		c, clk := newContext()

		// Initial value.
		s := Slot{
			Fetcher: func(c context.Context, prev Value) (Value, error) {
				return Value{1, clk.Now().Add(time.Second)}, nil
			},
		}
		v, err := s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 1)

		// Make it expire. Start panicing fetch.
		clk.Add(5 * time.Second)
		s.Fetcher = func(c context.Context, prev Value) (Value, error) {
			panic("omg")
		}
		So(func() { s.Get(c) }, ShouldPanicWith, "omg")

		// Doesn't deadlock.
		s.Fetcher = func(c context.Context, prev Value) (Value, error) {
			return Value{2, clk.Now().Add(time.Second)}, nil
		}
		v, err = s.Get(c)
		So(err, ShouldBeNil)
		So(v.Value.(int), ShouldEqual, 2)
	})

	Convey("Checks for nil", t, func() {
		c, clk := newContext()
		s := Slot{
			Fetcher: func(c context.Context, prev Value) (Value, error) {
				return Value{nil, clk.Now().Add(time.Second)}, nil
			},
		}
		So(func() { s.Get(c) }, ShouldPanicWith, "lazyslot.Slot Fetcher returned nil value")
	})

	Convey("Retries failed refetch later", t, func() {
		c, clk := newContext()

		var errorToReturn error
		var valueToReturn int

		fetchCalls := 0

		s := Slot{
			Fetcher: func(c context.Context, prev Value) (Value, error) {
				fetchCalls++
				return Value{
					Value:      valueToReturn,
					Expiration: clk.Now().Add(time.Minute),
				}, errorToReturn
			},
		}

		// Initial fetch.
		valueToReturn = 1
		errorToReturn = nil
		val, err := s.Get(c)
		So(val.Value, ShouldResemble, 1)
		So(err, ShouldBeNil)
		So(fetchCalls, ShouldEqual, 1)

		// Cached copy is good after 30 sec.
		clk.Add(30 * time.Second)
		valueToReturn = 2
		errorToReturn = nil
		val, err = s.Get(c)
		So(val.Value, ShouldResemble, 1) // still cached copy
		So(err, ShouldBeNil)
		So(fetchCalls, ShouldEqual, 1)

		// After 31 the cache copy expires, we attempt to update it, but something
		// goes horribly wrong. Get(...) returns the old copy, with expiration time
		// bumped to now+5 sec.
		clk.Add(31 * time.Second)
		valueToReturn = 3
		errorToReturn = errors.New("omg")
		val, err = s.Get(c)
		So(val.Value, ShouldResemble, 1) // still cached copy
		So(val.Expiration, ShouldResemble, clk.Now().Add(5*time.Second))
		So(err, ShouldBeNil)
		So(fetchCalls, ShouldEqual, 2) // attempted to fetch

		// 1 sec later still using old copy, because retry is scheduled for later.
		clk.Add(time.Second)
		valueToReturn = 4
		errorToReturn = nil
		val, err = s.Get(c)
		So(val.Value, ShouldResemble, 1) // still cached copy
		So(err, ShouldBeNil)
		So(fetchCalls, ShouldEqual, 2)

		// 5 sec later fetched is attempted, and it succeeds.
		clk.Add(5 * time.Second)
		valueToReturn = 5
		errorToReturn = nil
		val, err = s.Get(c)
		So(val.Value, ShouldResemble, 5) // new copy
		So(err, ShouldBeNil)
		So(fetchCalls, ShouldEqual, 3)
	})
}

func newContext() (context.Context, testclock.TestClock) {
	return testclock.UseTime(context.Background(), time.Unix(1442270520, 0))
}
