// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	cloudBQ "cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kartepb "infra/cros/karte/api"
	"infra/cros/karte/internal/identifiers"
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
	start := identifiers.IDInfo{
		Version:        req.GetStartVersion(),
		CoarseTime:     uint64(req.GetStartTime().GetSeconds()),
		FineTime:       uint32(req.GetStartTime().GetNanos()),
		Disambiguation: 0,
	}

	stop := identifiers.IDInfo{
		Version:        req.GetStopVersion(),
		CoarseTime:     uint64(req.GetStopTime().GetSeconds()),
		FineTime:       uint32(req.GetStopTime().GetNanos()),
		Disambiguation: 0,
	}

	tally, err := persistActionRangeImpl(ctx, &actionRangePersistOptions{
		startID: start,
		stopID:  stop,
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
