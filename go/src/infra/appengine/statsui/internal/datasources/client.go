// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package datasources

import (
	"context"
	"errors"
	"math/big"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"google.golang.org/api/iterator"
	"infra/appengine/statsui/internal/model"
)

type Client struct {
	Client *bigquery.Client
	Config *Config
}

var DataSourceNotFound = errors.New("data source not found")
var PeriodNotAvailable = errors.New("period not available")

func bqToDateArray(dates []string) ([]civil.Date, error) {
	ret := make([]civil.Date, len(dates))
	for i, date := range dates {
		d, err := civil.ParseDate(date)
		if err != nil {
			return nil, err
		}
		ret[i] = d
	}
	return ret, nil
}

func (c *Client) GetMetrics(ctx context.Context, dataSource string, period model.Period, dates []string, metrics []string) (map[string][]*model.Metric, error) {
	if _, exists := c.Config.Sources[dataSource]; !exists {
		return nil, DataSourceNotFound
	}
	if _, exists := c.Config.Sources[dataSource].Queries[period]; !exists {
		return nil, PeriodNotAvailable
	}
	type LabelValue struct {
		Label string
		Value *big.Rat
	}
	type Row struct {
		Date      civil.Date
		Section   string
		Metric    string
		Value     *big.Rat
		Aggregate []LabelValue
	}
	q := c.Client.Query(c.Config.Sources[dataSource].Queries[period])
	bqDates, err := bqToDateArray(dates)
	if err != nil {
		return nil, err
	}
	q.Parameters = []bigquery.QueryParameter{
		{Name: "dates", Value: bqDates},
		{Name: "metrics", Value: metrics},
	}
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	data := make(map[string]map[string]*model.Metric)
	for {
		var row Row
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		section, ok := data[row.Section]
		if !ok {
			section = make(map[string]*model.Metric)
			data[row.Section] = section
		}
		metric, ok := section[row.Metric]
		if !ok {
			metric = &model.Metric{
				Name:     row.Metric,
				Data:     make(model.DataSet),
				Sections: make(map[string]model.DataSet),
			}
			section[metric.Name] = metric
		}
		if row.Aggregate != nil {
			for _, val := range row.Aggregate {
				data, ok := metric.Sections[val.Label]
				if !ok {
					data = make(model.DataSet)
					metric.Sections[val.Label] = data
				}
				data[row.Date.String()], _ = val.Value.Float32()
			}
		} else if row.Value != nil {
			metric.Data[row.Date.String()], _ = row.Value.Float32()
		}
	}

	ret := make(map[string][]*model.Metric)
	for sectionName, m := range data {
		for _, metric := range m {
			ret[sectionName] = append(ret[sectionName], metric)
		}
	}

	return ret, nil
}
