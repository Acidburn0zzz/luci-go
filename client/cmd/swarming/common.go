// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"errors"
	"os"
	"runtime"

	"github.com/luci/luci-go/client/internal/common"
	"github.com/luci/luci-go/client/internal/lhttp"
	"github.com/maruel/subcommands"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type commonFlags struct {
	subcommands.CommandRunBase
	defaultFlags common.Flags
	serverURL    string
}

// Init initializes common flags.
func (c *commonFlags) Init() {
	c.defaultFlags.Init(&c.Flags)
	c.Flags.StringVar(&c.serverURL, "server", os.Getenv("SWARMING_SERVER"), "Server URL; required. Set $SWARMING_SERVER to set a default.")
}

// Parse parses the common flags.
func (c *commonFlags) Parse(a subcommands.Application) error {
	if err := c.defaultFlags.Parse(); err != nil {
		return err
	}
	if c.serverURL == "" {
		return errors.New("must provide -server")
	}
	s, err := lhttp.CheckURL(c.serverURL)
	if err != nil {
		return err
	}
	c.serverURL = s
	return nil
}
