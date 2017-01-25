// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ui

import (
	"github.com/luci/luci-go/server/router"
	"github.com/luci/luci-go/server/templates"
)

func projectPage(c *router.Context) {
	projectID := c.Params.ByName("ProjectID")
	jobs, err := config(c.Context).Engine.GetProjectJobs(c.Context, projectID)
	if err != nil {
		panic(err)
	}
	templates.MustRender(c.Context, c.Writer, "pages/project.html", map[string]interface{}{
		"ProjectID": projectID,
		"Jobs":      sortJobs(c.Context, jobs),
	})
}
