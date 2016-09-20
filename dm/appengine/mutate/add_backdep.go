// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mutate

import (
	ds "github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/dm/appengine/model"
	"github.com/luci/luci-go/tumble"
	"golang.org/x/net/context"
)

// AddBackDep adds a BackDep (and possibly a BackDepGroup). If NeedsAck
// is true, this mutation will chain to an AckFwdDep. It should only be false
// if this AddBackDep is spawned from an AddFinishedDeps, where the originating
// Attempt already knows that this dependency is Finished.
type AddBackDep struct {
	Dep      *model.FwdEdge
	NeedsAck bool // make AckFwdDep iff true
}

// Root implements tumble.Mutation.
func (a *AddBackDep) Root(c context.Context) *ds.Key {
	bdg, _ := a.Dep.Back(c)
	return ds.KeyForObj(c, bdg)
}

// RollForward implements tumble.Mutation.
func (a *AddBackDep) RollForward(c context.Context) (muts []tumble.Mutation, err error) {
	bdg, bd := a.Dep.Back(c)
	err = ds.Get(c, bdg)
	if err == ds.ErrNoSuchEntity {
		err = ds.Put(c, bdg)
	}
	if err != nil {
		return
	}

	bd.Propagated = bdg.AttemptFinished

	err = ds.Put(c, bd)

	if a.NeedsAck && bdg.AttemptFinished {
		muts = append(muts, &AckFwdDep{a.Dep})
	}
	return
}

func init() {
	tumble.Register((*AddBackDep)(nil))
}
