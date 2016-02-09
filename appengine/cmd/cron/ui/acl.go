// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ui

import (
	"golang.org/x/net/context"

	"github.com/luci/luci-go/server/auth"
)

func isJobOwner(c context.Context, projectID, jobID string) bool {
	// TODO(vadimsh): Do real ACLs.
	ok, err := auth.IsMember(c, "administrators")
	if err != nil {
		panic(err)
	}
	return ok
}
