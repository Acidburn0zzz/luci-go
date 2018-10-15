// Copyright 2018 The LUCI Authors.
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

package pkg

import (
	"context"
	"io"
)

// Source is an underlying data source with CIPD package data.
type Source interface {
	io.ReadSeeker

	// Close can be used to indicate to the storage (filesystem and/or cache)
	// layer that this instance is actually bad. The storage layer can then
	// evict/revoke, etc. the bad file.
	Close(ctx context.Context, corrupt bool) error
}
