// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package target

import (
	"github.com/luci/luci-go/common/tsmon/types"
	"golang.org/x/net/context"
)

type key int

const targetKey = iota

// Set returns a new context with the given target set.  If this context is
// passed to metric Set, Get or Incr methods the metrics for that target will be
// affected.  A nil target means to use the default target.
func Set(ctx context.Context, t types.Target) context.Context {
	return context.WithValue(ctx, targetKey, t)
}

// Get returns the target set in this context.
func Get(ctx context.Context) types.Target {
	if t, ok := ctx.Value(targetKey).(types.Target); ok {
		return t
	}
	return nil
}
