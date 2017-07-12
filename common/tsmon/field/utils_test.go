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

package field

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func makeInterfaceSlice(v ...interface{}) []interface{} {
	return v
}

func TestCanonicalize(t *testing.T) {
	data := []struct {
		fields []Field
		values []interface{}
		want   []interface{}
	}{
		{
			fields: []Field{},
			values: []interface{}{},
			want:   []interface{}{},
		},
		{
			fields: []Field{},
			values: makeInterfaceSlice(123),
			want:   nil,
		},
		{
			fields: []Field{String("foo")},
			values: makeInterfaceSlice(),
			want:   nil,
		},
		{
			fields: []Field{String("foo")},
			values: makeInterfaceSlice("v"),
			want:   makeInterfaceSlice("v"),
		},
		{
			fields: []Field{String("foo")},
			values: makeInterfaceSlice(123),
			want:   nil,
		},
		{
			fields: []Field{String("foo")},
			values: makeInterfaceSlice(true),
			want:   nil,
		},
		{
			fields: []Field{Int("foo")},
			values: makeInterfaceSlice(),
			want:   nil,
		},
		{
			fields: []Field{Int("foo")},
			values: makeInterfaceSlice("v"),
			want:   nil,
		},
		{
			fields: []Field{Int("foo")},
			values: makeInterfaceSlice(int(123)),
			want:   makeInterfaceSlice(int64(123)),
		},
		{
			fields: []Field{Int("foo")},
			values: makeInterfaceSlice(int32(123)),
			want:   makeInterfaceSlice(int64(123)),
		},
		{
			fields: []Field{Int("foo")},
			values: makeInterfaceSlice(int64(123)),
			want:   makeInterfaceSlice(int64(123)),
		},
		{
			fields: []Field{Int("foo")},
			values: makeInterfaceSlice(true),
			want:   nil,
		},
		{
			fields: []Field{Bool("foo")},
			values: makeInterfaceSlice(),
			want:   nil,
		},
		{
			fields: []Field{Bool("foo")},
			values: makeInterfaceSlice("v"),
			want:   nil,
		},
		{
			fields: []Field{Bool("foo")},
			values: makeInterfaceSlice(123),
			want:   nil,
		},
		{
			fields: []Field{Bool("foo")},
			values: makeInterfaceSlice(true),
			want:   makeInterfaceSlice(true),
		},
	}

	for i, d := range data {
		Convey(fmt.Sprintf("%d. Canonicalize(%v, %v)", i, d.fields, d.values), t, func() {
			ret, err := Canonicalize(d.fields, d.values)

			if d.want == nil {
				So(ret, ShouldBeNil)
				So(err, ShouldNotBeNil)
			} else {
				So(ret, ShouldResemble, d.want)
				So(err, ShouldBeNil)
			}
		})
	}
}

func TestHash(t *testing.T) {
	Convey("Empty slice hashes to 0", t, func() {
		So(Hash([]interface{}{}), ShouldEqual, 0)
	})

	Convey("Different things have different hashes", t, func() {
		hashes := map[uint64]struct{}{}
		values := [][]interface{}{
			makeInterfaceSlice(int64(123)),
			makeInterfaceSlice(int64(456)),
			makeInterfaceSlice(int64(0x01000000000000)),
			makeInterfaceSlice(int64(0x02000000000000)),
			makeInterfaceSlice(int64(123), int64(456)),
			makeInterfaceSlice("foo"),
			makeInterfaceSlice("bar"),
			makeInterfaceSlice("foo", "bar"),
			makeInterfaceSlice(true),
			makeInterfaceSlice(false),
			makeInterfaceSlice(true, false),
		}

		for _, v := range values {
			h := Hash(v)
			_, ok := hashes[h]
			So(ok, ShouldBeFalse)
			hashes[h] = struct{}{}
		}
	})
}
