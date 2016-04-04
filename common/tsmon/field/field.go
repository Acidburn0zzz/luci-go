// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package field

import (
	"fmt"

	pb "github.com/luci/luci-go/common/tsmon/ts_mon_proto"
)

// Field is the definition of a metric field.  It has a name and a type
// (string, int or bool).
type Field struct {
	Name string
	Type pb.MetricsField_FieldType
}

func (f Field) String() string {
	return fmt.Sprintf("Field(%s, %s)", f.Name, f.Type)
}

// String returns a new string-typed field.
func String(name string) Field { return Field{name, pb.MetricsField_STRING} }

// Bool returns a new bool-typed field.
func Bool(name string) Field { return Field{name, pb.MetricsField_BOOL} }

// Int returns a new int-typed field.  Internally values for these fields are
// stored as int64s.
func Int(name string) Field { return Field{name, pb.MetricsField_INT} }

// Serialize returns a slice of ts_mon_proto.MetricsField messages representing
// the field names, types and values.
func Serialize(fields []Field, values []interface{}) []*pb.MetricsField {
	ret := make([]*pb.MetricsField, len(fields))

	for i, f := range fields {
		d := &pb.MetricsField{
			Name: f.Name,
			Type: f.Type,
		}

		switch f.Type {
		case pb.MetricsField_STRING:
			d.StringValue = values[i].(string)
		case pb.MetricsField_BOOL:
			d.BoolValue = values[i].(bool)
		case pb.MetricsField_INT:
			d.IntValue = values[i].(int64)
		}

		ret[i] = d
	}

	return ret
}
