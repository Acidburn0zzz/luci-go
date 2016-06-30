// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package swarming

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/luci/luci-go/common/clock/testclock"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

var generate = flag.Bool("test.generate", false, "Generate expectations instead of running tests.")

func load(name string) ([]byte, error) {
	filename := path.Join("expectations", name)
	return ioutil.ReadFile(filename)
}

func shouldMatchExpectationsFor(actualContents interface{}, expectedFilename ...interface{}) string {
	refBuild, err := load(expectedFilename[0].(string))
	if err != nil {
		return fmt.Sprintf("Could not load %s: %s", expectedFilename[0], err.Error())
	}
	actualBuild, err := json.MarshalIndent(actualContents, "", "  ")
	return ShouldEqual(string(actualBuild), string(refBuild))

}

func TestBuild(t *testing.T) {
	testCases := []struct {
		input        string
		expectations string
	}{
		{"build-exception", "build-exception.json"},
		{"build-patch-failure", "build-patch-failure.json"},
		{"build-pending", "build-pending.json"},
		{"build-running", "build-running.json"},
		{"build-timeout", "build-timeout.json"},
		{"build-canceled", "build-canceled.json"},
	}

	c := context.Background()
	c, _ = testclock.UseTime(c, time.Date(2016, time.March, 14, 11, 0, 0, 0, time.UTC))

	if *generate {
		for _, tc := range testCases {
			fmt.Printf("Generating expectations for %s\n", tc.input)
			build, err := swarmingBuildImpl(c, "foo", "debug", tc.input)
			if err != nil {
				panic(fmt.Errorf("Could not run swarmingBuildImpl for %s: %s", tc.input, err))
			}
			buildJSON, err := json.MarshalIndent(build, "", "  ")
			if err != nil {
				panic(fmt.Errorf("Could not JSON marshal %s: %s", tc.input, err))
			}
			filename := path.Join("expectations", tc.expectations)
			err = ioutil.WriteFile(filename, []byte(buildJSON), 0644)
			if err != nil {
				panic(fmt.Errorf("Encountered error while trying to write to %s: %s", filename, err))
			}
		}
		return
	}

	Convey(`A test Environment`, t, func() {
		for _, tc := range testCases {
			Convey(fmt.Sprintf("Test Case: %s", tc.input), func() {
				build, err := swarmingBuildImpl(c, "foo", "debug", tc.input)
				So(err, ShouldBeNil)
				So(build, shouldMatchExpectationsFor, tc.expectations)
			})
		}
	})
}
