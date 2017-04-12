// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package buildbot

import (
	"net/http"

	"github.com/luci/luci-go/appengine/gaemiddleware"
	"github.com/luci/luci-go/server/router"
)

func init() {
	r := router.New()

	gaemiddleware.InstallHandlers(r)
	InstallAPIRoutes(r, gaemiddleware.BaseProd())
	http.DefaultServeMux.Handle("/", r)
}
