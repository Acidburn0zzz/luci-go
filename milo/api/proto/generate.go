// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package milo

//go:generate go install github.com/luci/luci-go/grpc/cmd/cproto
//go:generate go install github.com/luci/luci-go/grpc/cmd/svcdec
//go:generate cproto
//go:generate svcdec -type BuildbotServer -type BuildInfoServer
