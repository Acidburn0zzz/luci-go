// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package authtest

import (
	"net"

	"golang.org/x/net/context"

	"github.com/luci/luci-go/server/auth"
	"github.com/luci/luci-go/server/auth/authdb"
	"github.com/luci/luci-go/server/auth/identity"
	"github.com/luci/luci-go/server/auth/signing"
)

// FakeDB implements user group checking part of db.DB (IsMember).
//
// It is a mapping "identity -> list of its groups". Intended to be used mostly
// for testing request handlers, thus all other DB methods (that used by auth
// system when authenticating the request) is not implement and panic when
// called: the wast majority of request handlers are not calling them.
type FakeDB map[identity.Identity][]string

var _ authdb.DB = (FakeDB)(nil)

// Use installs the fake db into the context.
func (db FakeDB) Use(c context.Context) context.Context {
	return auth.ModifyConfig(c, func(cfg auth.Config) auth.Config {
		cfg.DBProvider = func(context.Context) (authdb.DB, error) {
			return db, nil
		}
		return cfg
	})
}

// IsMember is part of authdb.DB interface.
//
// It returns true if any of 'groups' is listed in db[id].
func (db FakeDB) IsMember(c context.Context, id identity.Identity, groups ...string) (bool, error) {
	for _, group := range groups {
		for _, gr := range db[id] {
			if gr == group {
				return true, nil
			}
		}
	}
	return false, nil
}

// IsAllowedOAuthClientID is part of authdb.DB interface. Panics.
func (db FakeDB) IsAllowedOAuthClientID(c context.Context, email, clientID string) (bool, error) {
	panic("FakeDB.IsAllowedOAuthClientID must not be called")
}

// GetCertificates is part of authdb.DB interface. Panics.
func (db FakeDB) GetCertificates(c context.Context, id identity.Identity) (*signing.PublicCertificates, error) {
	panic("FakeDB.GetCertificates must not be called")
}

// GetWhitelistForIdentity is part of authdb.DB interface. Panics.
func (db FakeDB) GetWhitelistForIdentity(c context.Context, ident identity.Identity) (string, error) {
	panic("FakeDB.GetWhitelistForIdentity must not be called")
}

// IsInWhitelist is part of authdb.DB interface. Panics.
func (db FakeDB) IsInWhitelist(c context.Context, ip net.IP, whitelist string) (bool, error) {
	panic("FakeDB.IsInWhitelist must not be called")
}

// GetAuthServiceURL is part of authdb.DB interface. Panics.
func (db FakeDB) GetAuthServiceURL(c context.Context) (string, error) {
	panic("FakeDB.GetAuthServiceURL must not be called")
}

// GetTokenServiceURL is part of authdb.DB interface. Panics.
func (db FakeDB) GetTokenServiceURL(c context.Context) (string, error) {
	panic("FakeDB.GetTokenServiceURL must not be called")
}

// FakeErroringDB is authdb.DB with IsMember returning an error.
type FakeErroringDB struct {
	FakeDB

	// Error is returned by IsMember.
	Error error
}

// IsMember is part of authdb.DB interface.
//
// It returns db.Error if it is not nil.
func (db *FakeErroringDB) IsMember(c context.Context, id identity.Identity, groups ...string) (bool, error) {
	if db.Error != nil {
		return false, db.Error
	}
	return db.FakeDB.IsMember(c, id, groups...)
}

// Use installs the fake db into the context.
func (db *FakeErroringDB) Use(c context.Context) context.Context {
	return auth.ModifyConfig(c, func(cfg auth.Config) auth.Config {
		cfg.DBProvider = func(context.Context) (authdb.DB, error) {
			return db, nil
		}
		return cfg
	})
}
