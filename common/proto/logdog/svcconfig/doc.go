// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package svcconfig stores service configuration for a LogDog instance.
//
// Each LogDog instantiation will have a single Config protobuf. It will be
// located under config set "services/<app-id>", path "services.cfg". The path
// is exposed via ServiceConfigFilename.
//
// Each LogDog project will have its own project-specific configuration. It will
// be located under config set "projects/<project-name>", path "<app-id>.cfg".
package svcconfig
