// Copyright 2015 The LUCI Authors.
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

// Package ui implements request handlers that serve user facing HTML pages.
package ui

import (
	"strings"

	"golang.org/x/net/context"

	"github.com/luci/gae/service/info"

	"github.com/luci/luci-go/appengine/gaeauth/server"
	"github.com/luci/luci-go/server/auth"
	"github.com/luci/luci-go/server/auth/xsrf"
	"github.com/luci/luci-go/server/router"
	"github.com/luci/luci-go/server/templates"

	"github.com/luci/luci-go/scheduler/appengine/catalog"
	"github.com/luci/luci-go/scheduler/appengine/engine"
)

// Config is global configuration of UI handlers.
type Config struct {
	Engine        engine.Engine
	Catalog       catalog.Catalog
	TemplatesPath string // path to templates directory deployed to GAE
}

// InstallHandlers adds HTTP handlers that render HTML pages.
func InstallHandlers(r *router.Router, base router.MiddlewareChain, cfg Config) {
	tmpl := prepareTemplates(cfg.TemplatesPath)

	m := base.Extend(func(c *router.Context, next router.Handler) {
		c.Context = context.WithValue(c.Context, configContextKey(0), &cfg)
		next(c)
	})
	m = m.Extend(
		templates.WithTemplates(tmpl),
		auth.Authenticate(server.UsersAPIAuthMethod{}),
	)

	r.GET("/", m, indexPage)
	r.GET("/jobs/:ProjectID", m, projectPage)
	r.GET("/jobs/:ProjectID/:JobName", m, jobPage)
	r.GET("/jobs/:ProjectID/:JobName/:InvID", m, invocationPage)

	// All POST forms must be protected with XSRF token.
	mxsrf := m.Extend(xsrf.WithTokenCheck)
	r.POST("/actions/runJob/:ProjectID/:JobName", mxsrf, runJobAction)
	r.POST("/actions/pauseJob/:ProjectID/:JobName", mxsrf, pauseJobAction)
	r.POST("/actions/resumeJob/:ProjectID/:JobName", mxsrf, resumeJobAction)
	r.POST("/actions/abortJob/:ProjectID/:JobName", mxsrf, abortJobAction)
	r.POST("/actions/abortInvocation/:ProjectID/:JobName/:InvID", mxsrf, abortInvocationAction)
}

type configContextKey int

// config returns Config passed to InstallHandlers.
func config(c context.Context) *Config {
	cfg, _ := c.Value(configContextKey(0)).(*Config)
	if cfg == nil {
		panic("impossible, configContextKey is not set")
	}
	return cfg
}

// prepareTemplates configures templates.Bundle used by all UI handlers.
//
// In particular it includes a set of default arguments passed to all templates.
func prepareTemplates(templatesPath string) *templates.Bundle {
	return &templates.Bundle{
		Loader:          templates.FileSystemLoader(templatesPath),
		DebugMode:       info.IsDevAppServer,
		DefaultTemplate: "base",
		DefaultArgs: func(c context.Context) (templates.Args, error) {
			loginURL, err := auth.LoginURL(c, "/")
			if err != nil {
				return nil, err
			}
			logoutURL, err := auth.LogoutURL(c, "/")
			if err != nil {
				return nil, err
			}
			token, err := xsrf.Token(c)
			if err != nil {
				return nil, err
			}
			return templates.Args{
				"AppVersion":  strings.Split(info.VersionID(c), ".")[0],
				"IsAnonymous": auth.CurrentIdentity(c) == "anonymous:anonymous",
				"User":        auth.CurrentUser(c),
				"LoginURL":    loginURL,
				"LogoutURL":   logoutURL,
				"XsrfToken":   token,
			}, nil
		},
	}
}
