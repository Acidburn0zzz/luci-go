// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package isolate implements the code to process '.isolate' files.
package isolate

const ISOLATED_GEN_JSON_VERSION = 1

// Tree to be isolated.
type Tree struct {
	Cwd  string
	Opts ArchiveOptions
}

// ArchiveOptions for achiving trees.
type ArchiveOptions struct {
	Isolate         string            `json:"isolate"`
	Isolated        string            `json:"isolated"`
	Blacklist       []string          `json:"blacklist"`
	PathVariables   map[string]string `json:"path_variables"`
	ExtraVariables  map[string]string `json:"extra_variables"`
	ConfigVariables map[string]string `json:"config_variables"`
}

// NewArchiveOptions initializes with non-nil values
func (a *ArchiveOptions) Init() {
	a.Blacklist = []string{}
	a.PathVariables = map[string]string{}
	a.ExtraVariables = map[string]string{}
	a.ConfigVariables = map[string]string{}
}

func IsolateAndArchive(trees []Tree, namespace string, server string) (
	map[string]string, error) {
	return nil, nil
}
