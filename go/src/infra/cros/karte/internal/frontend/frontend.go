// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"time"

	cloudBQ "cloud.google.com/go/bigquery"
	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kartepb "infra/cros/karte/api"
	"infra/cros/karte/internal/datastore"
	"infra/cros/karte/internal/errors"
	"infra/cros/karte/internal/idserialize"
	"infra/cros/karte/internal/scalars"
)

// karteFrontend is the implementation of kartepb.KarteServer
// used in the application.
type karteFrontend struct{}

// KarteFrontend is a combination of the Karte RPCs and the cron RPCs.
// In the future, any other services exposes by Karte should also be added here.
type KarteFrontend interface {
	kartepb.KarteServer
	kartepb.KarteCronServer
}

// NewKarteFrontend produces a new Karte frontend.
func NewKarteFrontend() KarteFrontend {
	return &karteFrontend{}
}

// PersistAction persists a single action.
func (k *karteFrontend) PersistAction(ctx context.Context, req *kartepb.PersistActionRequest) (*kartepb.PersistActionResponse, error) {
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
func (k *karteFrontend) persistActionRangeImpl(ctx context.Context, client bqPersister, req *kartepb.PersistActionRangeRequest) (*kartepb.PersistActionRangeResponse, error) {
	start := idserialize.IDInfo{
		Version:        req.GetStartVersion(),
		CoarseTime:     uint64(req.GetStartTime().GetSeconds()),
		FineTime:       uint32(req.GetStartTime().GetNanos()),
		Disambiguation: 0,
	}

	stop := idserialize.IDInfo{
		Version:        req.GetStopVersion(),
		CoarseTime:     uint64(req.GetStopTime().GetSeconds()),
		FineTime:       uint32(req.GetStopTime().GetNanos()),
		Disambiguation: 0,
	}

	q, err := newActionNameRangeQuery(start, stop)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "persist action range impl: failed to build query: %s", err)
	}

	const stride = 1000

	logging.Infof(ctx, "Beginning to insert record to bigquery")

	// TODO(gregorynisbet): This function doesn't need to exist.
	//                      Remove this function eventually as part of using ValueSaver everywhere.
	insertCb := func(ctx context.Context, ents []*ActionEntity) error {
		valueSavers := make([]cloudBQ.ValueSaver, 0, len(ents))
		for _, ent := range ents {
			valueSavers = append(valueSavers, ent.ConvertToValueSaver())
		}
		f := client.getInserter("entities", "actions")
		err := f(ctx, valueSavers)
		return errors.Annotate(err, "insert rows").Err()
	}

	tally := 0

	for q.Token != stopToken {
		batch, _, err := q.Next(ctx, stride)
		if err != nil {
			return nil, errors.Annotate(err, "persist action range").Err()
		}

		tally += len(batch)

		// TODO(gregorynisbet): A batch length of zero signals the successful end of the offload attempt.
		//                      Replace this with a better API for next.
		if len(batch) == 0 {
			break
		}

		// We do not need to worry about retries. The default implementation will retry for us in a reasonable way.
		if err := insertCb(ctx, batch); err != nil {
			logging.Errorf(ctx, "cannot insert action: %s", err)
			return nil, status.Errorf(codes.Aborted, "error persisting single record: %s", err)
		}
	}

	return &kartepb.PersistActionRangeResponse{
		Succeeded:      true,
		CreatedRecords: int32(tally),
	}, nil
}

// ListActions lists the actions that Karte knows about.
func (k *karteFrontend) ListActions(ctx context.Context, req *kartepb.ListActionsRequest) (*kartepb.ListActionsResponse, error) {
	q, err := newActionEntitiesQuery(req.GetPageToken(), req.GetFilter())
	if err != nil {
		return nil, errors.Annotate(err, "list actions").Err()
	}

	es, _, err := q.Next(ctx, req.GetPageSize())
	if err != nil {
		return nil, errors.Annotate(err, "list actions (page size: %d)", req.GetPageSize()).Err()
	}
	var actions []*kartepb.Action
	for _, e := range es {
		actions = append(actions, e.ConvertToAction())
	}
	return &kartepb.ListActionsResponse{
		Actions:       actions,
		NextPageToken: q.Token,
	}, nil
}

// ListObservations lists the observations that Karte knows about.
func (k *karteFrontend) ListObservations(ctx context.Context, req *kartepb.ListObservationsRequest) (*kartepb.ListObservationsResponse, error) {
	q, err := newObservationEntitiesQuery(req.GetPageToken(), req.GetFilter())
	if err != nil {
		return nil, errors.Annotate(err, "list observations").Err()
	}
	es, err := q.Next(ctx, req.GetPageSize())
	if err != nil {
		return nil, errors.Annotate(err, "list observations (page size: %d)", req.GetPageSize()).Err()
	}
	var observations []*kartepb.Observation
	for _, e := range es {
		observations = append(observations, e.ConvertToObservation())
	}
	return &kartepb.ListObservationsResponse{
		Observations:  observations,
		NextPageToken: q.Token,
	}, nil
}

// UpdateAction updates an action in datastore and creates it if necessary when allow_missing is set.
func (k *karteFrontend) UpdateAction(ctx context.Context, req *kartepb.UpdateActionRequest) (*kartepb.Action, error) {
	// TODO(gregorynisbet): Remove json logging.
	str, mErr := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(req)
	if mErr == nil {
		logging.Infof(ctx, "Update action request: %s", str)
	} else {
		logging.Errorf(ctx, "Failed to marshal action request: %s", mErr)
	}
	reqActionEntity, err := convertActionToActionEntity(req.GetAction())
	if err != nil {
		return nil, errors.Annotate(err, "update action").Err()
	}
	entity, err := UpdateActionEntity(
		ctx,
		reqActionEntity,
		req.GetUpdateMask().GetPaths(),
		true,
	)
	return entity.ConvertToAction(), err
}

// PersistToBigquery persists all Karte-tracked records in a given time range to BigQuery.
// This is a cron method and part of the cron group of API calls.
// It is intentionally EXACTLY equivalent to calling the non-cron API persist-action-range with
// "reasonable" arguments.
func (k *karteFrontend) PersistToBigquery(ctx context.Context, req *kartepb.PersistToBigqueryRequest) (*kartepb.PersistToBigqueryResponse, error) {
	now := time.Now()
	resp, err := k.PersistActionRange(
		ctx,
		&kartepb.PersistActionRangeRequest{
			// Look twelve hours into the past. A karte record is sealed after 12 hours, so there's no way
			// for us to miss an important update this way.
			//
			// Also, there will be duplicate records in the bq table this way but that's okay.
			StartTime: scalars.ConvertTimeToTimestampPtr(now.Add(-12 * time.Hour)),
			// Give ourselves some buffer and actually persist stuff that was created up to an hour in the future.
			StopTime: scalars.ConvertTimeToTimestampPtr(now.Add(+1 * time.Hour)),
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "persist to bigquery failed: %s", err)
	}
	return &kartepb.PersistToBigqueryResponse{
		Succeeded:      true,
		CreatedActions: resp.GetCreatedRecords(),
	}, nil
}

// InstallServices takes a Karte frontend and exposes it to a LUCI prpc.Server.
func InstallServices(srv *prpc.Server) {
	kartepb.RegisterKarteServer(srv, NewKarteFrontend())
	kartepb.RegisterKarteCronServer(srv, NewKarteFrontend())
}
