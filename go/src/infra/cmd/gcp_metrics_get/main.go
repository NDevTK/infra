// Copyright 2022 The LUCI Authors.
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

// Binary gcp_metrics_get retrieves metrics data from stackdriver.
//
// Usage:
//
//	$ gcp_metrics_get --project $PROJECT_ID --filter '...' --end_time YYYY-MM-DDThh:mm:dd.ss
//
//	$ gcp_metrics_get --project_id goma-rbe-chromium --filter \
//	 'metric.type="kubernetes.io/container/memory/request_utilization"
//	  resource.labels.container_name="esp"' | \
//	 jq --slurp -r 'sort_by(.points[0].value.Value.DoubleValue) |
//	       reverse | .[] | \
//	       select(.points[0].value.Value.DoubleValue >= 0.5) | \
//	       [.resource.labels.pod_name, .points[0].value.Value.DoubleValue] | \
//	       @tsv'
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	projectID          = flag.String("project_id", "", "Project ID")
	filter             = flag.String("filter", "", "metrics filter.  see https://cloud.google.com/monitoring/custom-metrics/reading-metrics#time-series_filters")
	endTimeStr         = flag.String("end_time", "now", fmt.Sprintf("end time in RFC3339 e.g. %s, or now", time.RFC3339))
	duration           = flag.Duration("duration", 10*time.Minute, "duration")
	headerOnly         = flag.Bool("header_only", false, "Returns the identity of the metric and the time series resource, but not the time series data.")
	aggregationFields  = flag.String("aggregation_fields", "", "fields used for group by aggregation, use ',' for multiple fields separation")
	aggregationReducer = flag.String("aggregation_reducer", "REDUCE_NONE", fmt.Sprintf("reducer for aggregation, possible values are %s", func() string {
		var reducers []string
		for key := range monitoringpb.Aggregation_Reducer_value {
			reducers = append(reducers, key)
		}
		sort.Strings(reducers)
		return strings.Join(reducers, ",")
	}()))
)

func main() {
	flag.Parse()
	if *projectID == "" {
		fmt.Fprintln(flag.CommandLine.Output(), "missing project_id")
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()

	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		log.Fatalf("failed to create metric client: %v", err)
	}
	defer c.Close()

	endTime := time.Now()
	if *endTimeStr != "" && *endTimeStr != "now" {
		endTime, err = time.Parse(time.RFC3339, *endTimeStr)
		if err != nil {
			log.Fatalf("bad endtime %q: %v", *endTimeStr, err)
		}
	}
	startTime := endTime.Add(-*duration)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + *projectID,
		Filter: *filter,
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
	}

	if *headerOnly {
		req.View = monitoringpb.ListTimeSeriesRequest_HEADERS
	}

	if *aggregationFields != "" {
		reducer, ok := monitoringpb.Aggregation_Reducer_value[*aggregationReducer]
		if !ok {
			log.Fatalf("aggregation_reducer %s is not supported, see -help.", *aggregationReducer)
		}
		req.Aggregation = &monitoringpb.Aggregation{
			GroupByFields:      strings.Split(*aggregationFields, ","),
			CrossSeriesReducer: monitoringpb.Aggregation_Reducer(reducer),
		}
	}

	iter := c.ListTimeSeries(ctx, req)

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("could not read time series value, %v ", err)
		}
		b, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("failed to marshal json %s: %v", resp, err)
		}
		fmt.Printf("%s\n", b)
	}
}
