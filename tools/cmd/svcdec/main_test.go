// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	inputDir   = filepath.Join("..", "..", "internal", "svctool", "testdata")
	goldenFile = "testdata/s1server_dec.golden"
)

func TestMain(t *testing.T) {
	t.Parallel()

	Convey("svxdec", t, func() {
		tmpDir, err := ioutil.TempDir("", "")
		So(err, ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		run := func(args ...string) error {
			t := tool()
			t.ParseArgs(args)
			return t.Run(context.Background(), generate)
		}

		Convey("Works", func() {
			output := filepath.Join(tmpDir, "s1server_dec.go")
			err := run(
				"-output", output,
				"-type", "S1Server,S2Server",
				inputDir,
			)
			So(err, ShouldBeNil)

			want, err := ioutil.ReadFile(goldenFile)
			So(err, ShouldBeNil)

			got, err := ioutil.ReadFile(output)
			So(err, ShouldBeNil)

			So(string(got), ShouldEqual, string(want))
		})
	})
}
