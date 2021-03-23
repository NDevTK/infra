// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"io/ioutil"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"
	"infra/appengine/statsui/api"
	"infra/appengine/statsui/internal/datasources"
	"infra/appengine/statsui/internal/model"
)

func main() {
	modules := []module.Module{
		cfgmodule.NewModuleFromFlags(),
		gaeemulation.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		dsClient, err := setupDataSourceClient(srv.Context)
		if err != nil {
			return err
		}
		stats := &statsServer{
			DataSources: dsClient,
		}
		api.RegisterStatsServer(srv.PRPC, stats)
		return nil
	})
}

func setupDataSourceClient(ctx context.Context) (*datasources.Client, error) {
	yaml, err := ioutil.ReadFile("datasources.yaml")
	if err != nil {
		return nil, err
	}
	config, err := datasources.LoadConfig(yaml)
	if err != nil {
		return nil, err
	}
	bqClient, err := bigquery.NewClient(ctx, "chrome-trooper-analytics")
	if err != nil {
		return nil, err
	}
	return &datasources.Client{
		Client: bqClient,
		Config: config,
	}, nil
}

type statsServer struct {
	DataSources *datasources.Client
}

func (s *statsServer) FetchMetrics(ctx context.Context, req *api.FetchMetricsRequest) (*api.FetchMetricsResponse, error) {
	var period model.Period
	switch req.Period {
	case api.Period_WEEK:
		period = model.Week
	case api.Period_DAY:
		period = model.Day
	default:
		return nil, errors.New("unsupported period")
	}
	metrics, err := s.DataSources.GetMetrics(ctx, req.Datasource, period, req.Dates, req.Metrics)
	if err != nil {
		return nil, err
	}
	resp := &api.FetchMetricsResponse{}
	for name, metrics := range metrics {
		section := &api.Section{Name: name}
		for _, metric := range metrics {
			section.Metrics = append(section.Metrics, toMetric(metric))
		}
		resp.Sections = append(resp.Sections, section)
	}
	return resp, nil
}

func toMetric(m *model.Metric) *api.Metric {
	ret := &api.Metric{Name: m.Name}
	if len(m.Data) != 0 {
		ret.Data = &api.DataSet{Data: m.Data}
	}
	if len(m.Sections) != 0 {
		ret.Sections = make(map[string]*api.DataSet)
		for name, data := range m.Sections {
			ret.Sections[name] = &api.DataSet{Data: data}
		}
	}
	return ret
}
