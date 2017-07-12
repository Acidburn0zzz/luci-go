// Copyright 2016 The LUCI Authors.
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

package dm

import (
	"fmt"
)

// validAttemptStateEvolution defines all valid {From -> []To} state
// transitions. The identity transition (X -> X) is implied, as long as X has an
// entry in this mapping.
var validAttemptStateEvolution = map[Attempt_State][]Attempt_State{
	Attempt_SCHEDULING: {
		Attempt_EXECUTING,         // scheduled
		Attempt_ABNORMAL_FINISHED, // cancelled/timeout/err/etc.
	},
	Attempt_EXECUTING: {
		Attempt_SCHEDULING,        // Retry
		Attempt_WAITING,           // EnsureGraphData
		Attempt_FINISHED,          // FinishAttempt
		Attempt_ABNORMAL_FINISHED, // cancel/timeout/err/etc.
	},
	Attempt_WAITING: {
		Attempt_SCHEDULING,        // unblocked
		Attempt_ABNORMAL_FINISHED, // cancelled
	},

	Attempt_FINISHED:          {},
	Attempt_ABNORMAL_FINISHED: {},
}

// Evolve attempts to evolve the state of this Attempt. If the state evolution
// is not allowed (e.g. invalid state transition), this returns an error.
func (s *Attempt_State) Evolve(newState Attempt_State) error {
	nextStates := validAttemptStateEvolution[*s]
	if nextStates == nil {
		return fmt.Errorf("invalid state transition: no transitions defined for %s", *s)
	}

	if newState == *s {
		return nil
	}

	for _, val := range nextStates {
		if newState == val {
			*s = newState
			return nil
		}
	}

	return fmt.Errorf("invalid state transition %v -> %v", *s, newState)
}

// MustEvolve is a panic'ing version of Evolve.
func (s *Attempt_State) MustEvolve(newState Attempt_State) {
	err := s.Evolve(newState)
	if err != nil {
		panic(err)
	}
}

// Terminal returns true iff there are no valid evolutions from the current
// state.
func (s Attempt_State) Terminal() bool {
	return len(validAttemptStateEvolution[s]) == 0
}
