// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package monitor

import (
	"os"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/logging"
	pb "github.com/luci/luci-go/common/tsmon/ts_mon_proto"
	"github.com/luci/luci-go/common/tsmon/types"
)

type debugMonitor struct {
	path string
}

// NewDebugMonitor returns a Monitor that outputs metrics to a log, and
// optionally a file on disk.
func NewDebugMonitor(path string) Monitor {
	return &debugMonitor{
		path: path,
	}
}

func (m *debugMonitor) ChunkSize() int {
	return 0
}

func (m *debugMonitor) Send(ctx context.Context, cells []types.Cell) error {
	str := proto.MarshalTextString(&pb.MetricsPayload{
		MetricsCollection: SerializeCells(cells, clock.Now(ctx)),
	})

	logging.Infof(ctx, "Sending ts_mon metrics:\n%s", str)

	if m.path != "" {
		file, err := os.OpenFile(m.path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
		if err != nil {
			logging.Errorf(ctx, "Failed to open file %s: %s", m.path, err)
			return err
		}

		defer func() {
			if err := file.Close(); err != nil {
				logging.Errorf(ctx, "Failed to close file %s: %s", m.path, err)
			}
		}()

		if _, err = file.WriteString(str); err != nil {
			logging.Errorf(ctx, "Failed to write to the file %s: %s", m.path, err)
			return err
		}
	}

	return nil
}

func (m *debugMonitor) Close() error {
	return nil
}
