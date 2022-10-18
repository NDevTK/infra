// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/grpcutil"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/cron"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kartepb "infra/cros/karte/api"
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
// It is intentionally EXACTLY equivalent to calling the non-cron API persist-action-range.
//
// By default, we take the current time, truncate it to the nearest hour, and set that to the END time.
// We then take one hour before the end time and make that the start time.
func (k *karteFrontend) PersistToBigquery(ctx context.Context, req *kartepb.PersistToBigqueryRequest) (*kartepb.PersistToBigqueryResponse, error) {
	now := time.Now().UTC()
	thisHour := now.Truncate(time.Hour)
	prevHour := thisHour.Add(-1 * time.Hour)
	resp, err := k.PersistActionRange(
		ctx,
		&kartepb.PersistActionRangeRequest{
			StartTime: scalars.ConvertTimeToTimestampPtr(prevHour),
			StopTime:  scalars.ConvertTimeToTimestampPtr(thisHour),
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
func InstallServices(k KarteFrontend, srv *prpc.Server) {
	kartepb.RegisterKarteServer(srv, k)
	kartepb.RegisterKarteCronServer(srv, k)
	cron.RegisterHandler(
		"persist-to-bq",
		func(ctx context.Context) error {
			_, err := k.PersistToBigquery(
				ctx,
				&kartepb.PersistToBigqueryRequest{},
			)
			err = grpcutil.WrapIfTransient(err)
			err = errors.Annotate(err, "persist to bq cron").Err()
			return err
		},
	)
}
