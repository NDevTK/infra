// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dashboard

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/luci/gae/service/info"
	"github.com/luci/luci-go/appengine/gaemiddleware"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/grpc/discovery"
	"github.com/luci/luci-go/grpc/prpc"
	"github.com/luci/luci-go/server/router"
	"github.com/luci/luci-go/server/templates"

	dashpb "infra/appengine/dashboard/api/dashboard"
	"infra/appengine/dashboard/backend"
)

var templateBundle = &templates.Bundle{
	Loader:    templates.FileSystemLoader("templates"),
	DebugMode: info.IsDevAppServer,
	FuncMap: template.FuncMap{
		"fmtDate": func(date time.Time) string {
			return date.Format("1-2-2006")
		},
	},
}

func pageBase() router.MiddlewareChain {
	return gaemiddleware.BaseProd().Extend(
		templates.WithTemplates(templateBundle),
	)
}

func init() {
	r := router.New()
	gaemiddleware.InstallHandlers(r)
	r.GET("/", pageBase(), dashboard)
	http.DefaultServeMux.Handle("/", r)

	var api prpc.Server
	dashpb.RegisterChopsServiceStatusServer(&api, &dashboardService{})
	discovery.Enable(&api)
	api.InstallHandlers(r, gaemiddleware.BaseProd())
}

// TemplateService bundles a backend.Service with its backend.ServiceIncident children.
type TemplateService struct {
	Service   backend.Service
	Incidents []backend.ServiceIncident
}

func createServicesPageData(c context.Context, after time.Time, before time.Time) (sla []TemplateService, nonSLA []TemplateService, err error) {
	services, e := backend.GetAllServices(c)
	if e != nil {
		logging.Errorf(c, "error getting Service entities %s", e)
		return nil, nil, e
	}

	for _, service := range services {
		closedQueryOpts := &backend.QueryOptions{
			After:  after,
			Before: before,
			Status: backend.IncidentStatusClosed,
		}
		closedIncidents, e := backend.GetServiceIncidents(c, service.ID, closedQueryOpts)
		if err != nil {
			logging.Errorf(c, "error getting ServiceIncident entities %s", e)
			return nil, nil, e
		}
		openQueryOpts := &backend.QueryOptions{
			Status: backend.IncidentStatusOpen,
		}
		openIncidents, e := backend.GetServiceIncidents(c, service.ID, openQueryOpts)
		if err != nil {
			logging.Errorf(c, "error getting ServiceIncident entities %s", e)
			return nil, nil, e
		}
		templateService := TemplateService{service, append(openIncidents, closedIncidents...)}
		if service.SLA == "" {
			nonSLA = append(nonSLA, templateService)
		} else {
			sla = append(sla, templateService)
		}
	}
	return

}

// dateEqual should be used to check if two time.Times occur within the same day.
// Using Time.Equal has too granular of equality.
func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func dashboard(ctx *router.Context) {
	c, w, r := ctx.Context, ctx.Writer, ctx.Request
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form",
			http.StatusInternalServerError)
		return
	}
	upto := r.Form.Get("upto")
	lastDate := time.Now()
	var newerDate int64
	if upto != "" {
		unixInt, err := strconv.ParseInt(upto, 10, 64)
		if err != nil {
			logging.Infof(c, "%v, %v", err, lastDate)
			http.Error(w, "failed to parse \"upto\" date paramater",
				http.StatusBadRequest)
			return
		}
		dateFromParams := time.Unix(unixInt, 0)
		if !dateEqual(lastDate, dateFromParams) {
			lastDate = dateFromParams
			newerDate = lastDate.AddDate(0, 0, 7).Unix()
		}
	}

	dates := []time.Time{}
	for i := 6; i >= 0; i-- {
		dates = append(dates, lastDate.AddDate(0, 0, -i))
	}

	sla, nonSLA, err := createServicesPageData(c, dates[0], lastDate)
	if err != nil {
		http.Error(w, "failed to create Services page data, see logs",
			http.StatusInternalServerError)
		return
	}

	templates.MustRender(c, w, "pages/dash.tmpl", templates.Args{
		"ChopsServices":  sla,
		"NonSLAServices": nonSLA,
		"Dates":          dates,
		"OlderDate":      dates[0].AddDate(0, 0, -1).Unix(),
		"NewerDate":      newerDate,
	})
}
