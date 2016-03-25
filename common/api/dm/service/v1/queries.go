// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dm

// AttemptListQuery returns a new GraphQuery for the given AttemptList.
func AttemptListQuery(fanout *AttemptList) *GraphQuery {
	return &GraphQuery{AttemptList: fanout}
}

// AttemptListQueryL returns a new GraphQuery for the given AttemptList
// literal.
func AttemptListQueryL(fanout map[string][]uint32) *GraphQuery {
	return &GraphQuery{AttemptList: NewAttemptList(fanout)}
}

// AttemptRangeQuery returns a new GraphQuery for the given AttemptRange
// specification.
func AttemptRangeQuery(quest string, low, high uint32) *GraphQuery {
	return &GraphQuery{
		AttemptRange: []*GraphQuery_AttemptRange{{quest, low, high}}}
}
