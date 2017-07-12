// Copyright 2014 The LUCI Authors.
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

package local

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReadManifest(t *testing.T) {
	var goodManifest = `{
  "format_version": "1",
  "package_name": "package/name"
}`

	Convey("readManifest can read valid manifest", t, func() {
		manifest, err := readManifest(strings.NewReader(goodManifest))
		So(manifest, ShouldResemble, Manifest{
			FormatVersion: "1",
			PackageName:   "package/name",
		})
		So(err, ShouldBeNil)
	})

	Convey("readManifest rejects invalid manifest", t, func() {
		manifest, err := readManifest(strings.NewReader("I'm not a manifest"))
		So(manifest, ShouldResemble, Manifest{})
		So(err, ShouldNotBeNil)
	})

	Convey("writeManifest works", t, func() {
		buf := &bytes.Buffer{}
		m := Manifest{
			FormatVersion: "1",
			PackageName:   "package/name",
		}
		So(writeManifest(&m, buf), ShouldBeNil)
		So(string(buf.Bytes()), ShouldEqual, goodManifest)
	})
}
