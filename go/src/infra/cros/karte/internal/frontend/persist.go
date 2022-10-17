// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"time"

	cloudBQ "cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kartepb "infra/cros/karte/api"
	"infra/cros/karte/internal/scalars"
)

// PersistAction persists a single action.
func (*karteFrontend) PersistAction(ctx context.Context, req *kartepb.PersistActionRequest) (*kartepb.PersistActionResponse, error) {
	client, err := cloudBQ.NewClient(ctx, cloudBQ.DetectProjectID)
	if err != nil {
		logging.Errorf(ctx, "Cannot create bigquery client: %s", err)
		return nil, status.Errorf(codes.Aborted, "persist action: cannot create bigquery client: %s", err)
	}
	id := req.GetActionId()
	if id == "" {
		logging.Errorf(ctx, "Cannot get action ID: %s", err)
		return nil, status.Errorf(codes.InvalidArgument, "persist action: request ID cannot be empty")
	}
	ent := ActionEntity{}
	ent.ID = id
	if err := datastore.Get(ctx, &ent); err != nil {
		logging.Errorf(ctx, "Cannot retrieve action: %s", err)
		return nil, errors.Annotate(err, "persist action").Err()
	}
	valueSaver := ent.ConvertToValueSaver()
	logging.Infof(ctx, "beginning to insert record to bigquery")
	tbl := client.Dataset("entities").Table("actions")
	inserter := tbl.Inserter()
	if err := inserter.Put(ctx, valueSaver); err != nil {
		logging.Errorf(ctx, "cannot insert action: %s", err)
		return nil, status.Errorf(codes.Aborted, "error persisting single record: %s", err)
	}

	return &kartepb.PersistActionResponse{
		Succeeded:     true,
		CreatedRecord: true,
	}, nil
}

// persistBqClient is a wrapper around the bigquery client that exposes only the interface necessary to persist to
// persist ranges of actions.
type persistBqClient struct {
	client *cloudBQ.Client
}

// bqInserter persists ranges of actions to BigQuery.
type bqInserter = func(context.Context, []cloudBQ.ValueSaver) error

// getInserter gets the inserter for a table in a dataset.
func (c persistBqClient) getInserter(dataset string, table string) bqInserter {
	return func(ctx context.Context, valueSavers []cloudBQ.ValueSaver) error {
		return c.client.Dataset(dataset).Table(table).Inserter().Put(ctx, valueSavers)
	}
}

// PersistActionRange persists a range of actions.
func (k *karteFrontend) PersistActionRange(ctx context.Context, req *kartepb.PersistActionRangeRequest) (*kartepb.PersistActionRangeResponse, error) {
	logging.Infof(ctx, "Persisting action range %v to bigquery.", req)
	client, err := cloudBQ.NewClient(ctx, cloudBQ.DetectProjectID)
	if err != nil {
		logging.Errorf(ctx, "Cannot create bigquery client: %s", err)
		return nil, status.Errorf(codes.Aborted, "persist action range: cannot create bigquery client: %s", err)
	}

	return k.persistActionRangeImpl(ctx, persistBqClient{client}, req)
}

// bqPersister is the subset of the BigQuery interface used by the implementation of persistAction.
type bqPersister interface {
	getInserter(dataset string, table string) bqInserter
}

// persistActionRangeImpl is the implementation of persist range action.
func (*karteFrontend) persistActionRangeImpl(ctx context.Context, client bqPersister, req *kartepb.PersistActionRangeRequest) (*kartepb.PersistActionRangeResponse, error) {
	if req.GetStartVersion() != "" && req.GetStartVersion() != "zzzz" {
		return nil, errors.Reason("unsupported version %q", req.GetStartVersion()).Err()
	}
	if req.GetStopVersion() != "" && req.GetStopVersion() != "zzzz" {
		return nil, errors.Reason("unsupported version %q", req.GetStopVersion()).Err()
	}

	tally, err := persistActionRangeImpl(ctx, &actionRangePersistOptions{
		startID: scalars.ConvertTimestampPtrToTime(req.GetStartTime()),
		stopID:  scalars.ConvertTimestampPtrToTime(req.GetStopTime()),
		bq:      client,
	})
	if err != nil {
		return nil, errors.Annotate(err, "persist action range impl").Err()
	}
	return &kartepb.PersistActionRangeResponse{
		Succeeded:      true,
		CreatedRecords: int32(tally),
	}, nil
}

// timeRangePair consists of a starting point and an ending point.
type timeRangePair struct {
	start time.Time
	stop  time.Time
}

// splitTimeRange takes a time range and a number of subranges to split it into and splits it.
func splitTimeRange(start time.Time, stop time.Time, entries int) ([]timeRangePair, error) {
	var out []timeRangePair
	duration, err := validateSplitTimeRange(start, stop, entries)
	if err != nil {
		return nil, err
	}

	for i := 0; i < entries; i++ {
		curStart := start.Add(scaleNanoseconds(duration.Nanoseconds(), float64(i)/float64(entries)))
		curStop := start.Add(scaleNanoseconds(duration.Nanoseconds(), float64(i+1)/float64(entries)))
		if i == entries-1 {
			curStop = stop
		}
		out = append(out, timeRangePair{start: curStart, stop: curStop})
	}

	return out, nil
}

func scaleNanoseconds(nanoseconds int64, scale float64) time.Duration {
	return time.Duration(float64(nanoseconds) * scale)
}

func validateSplitTimeRange(start time.Time, stop time.Time, entries int) (time.Duration, error) {
	switch {
	case start.Location() != time.UTC:
		return 0, errors.Reason("split time range: start time must be UTC not %q", start.Location().String()).Err()
	case stop.Location() != time.UTC:
		return 0, errors.Reason("split time range: stop time must be UTC not %q", start.Location().String()).Err()
	case start.Equal(stop):
		return 0, errors.Reason("split time range: start time cannot equal stop time %q", start.String()).Err()
	case start.After(stop):
		return 0, errors.Reason("split time range: start time cannot occur after stop %q", stop.String()).Err()
	case entries <= 0:
		return 0, errors.Reason("split time range: invalid number of entries %d", entries).Err()
	}
	duration := stop.Sub(start)
	switch {
	case duration.Microseconds() < 1:
		return 0, errors.Reason("split time range: start %q and stop %q must differ by at least one microsecond", start.String(), stop.String()).Err()
	case duration.Hours() > 30*60:
		return 0, errors.Reason("split time range: start %q and stop %q must differ by at most 30 days", start.String(), stop.String()).Err()
	}
	return duration, nil
}
