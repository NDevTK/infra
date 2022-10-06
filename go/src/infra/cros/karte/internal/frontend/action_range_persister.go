// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"

	cloudBQ "cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/cros/karte/internal/idserialize"
)

// actionRangePersistOptions is a structure that can be used to manage an attempt to persist a range of actions.
type actionRangePersistOptions struct {
	// startID is a structural representation of earliest Karte ID to persist to BigQuery.
	startID idserialize.IDInfo
	// stopID is a structural representation of the latest Karte ID to persist to BigQuery.
	stopID idserialize.IDInfo
	// bq is the client that we use to add ValueSavers to BigQuery tables.
	bq bqPersister
}

// run gathers up all the observations and actions and persists them.
func persistActionRangeImpl(ctx context.Context, a *actionRangePersistOptions) (int, error) {
	q, err := makeQuery(a)
	if err != nil {
		return 0, errors.Annotate(err, "run").Err()
	}
	ad, tally, err := persistActions(ctx, a, q)
	if err != nil {
		return 0, errors.Annotate(err, "run").Err()
	}
	if err := persistObservations(ctx, a, ad); err != nil {
		return 0, errors.Annotate(err, "run").Err()
	}
	return tally, nil
}

// makeQuery makes a query and attaches it to the persister.
func makeQuery(a *actionRangePersistOptions) (*ActionEntitiesQuery, error) {
	q, err := newActionNameRangeQuery(a.startID, a.stopID)
	if err != nil {
		return nil, errors.Annotate(err, "make query").Err()
	}
	return q, nil
}

// insertBatch inserts a batch of actions into BigQuery.
func insertBatch(ctx context.Context, a *actionRangePersistOptions, ents []*ActionEntity) error {
	if len(ents) == 0 {
		return nil
	}
	valueSavers := make([]cloudBQ.ValueSaver, 0, len(ents))
	// This conversion right here, in a perfect world, would not be necessary.
	// "Next" should just return an array of valuesavers, but that is a problem for another day.
	for _, ent := range ents {
		valueSavers = append(valueSavers, ent.ConvertToValueSaver())
	}
	f := a.bq.getInserter("entities", "actions")
	return errors.Annotate(f(ctx, valueSavers), "insert batch").Err()
}

// insertObservationBatch inserts a batch of observations into BigQuery.
func insertObservationBatch(ctx context.Context, a *actionRangePersistOptions, ents []*ObservationEntity) error {
	if len(ents) == 0 {
		return nil
	}
	valueSavers := make([]cloudBQ.ValueSaver, 0, len(ents))
	for _, ent := range ents {
		valueSavers = append(valueSavers, ent.ConvertToValueSaver())
	}
	f := a.bq.getInserter("entities", "observations")
	return errors.Annotate(f(ctx, valueSavers), "insert batch").Err()
}

// persistActions persists all the actions corresponding to our attached query to bigquery.
func persistActions(ctx context.Context, a *actionRangePersistOptions, q *ActionEntitiesQuery) (*ActionQueryAncillaryData, int, error) {
	out := &ActionQueryAncillaryData{}
	tally := 0
	for q.Token != stopToken {
		batch, ad, err := q.Next(ctx, defaultBatchSize)
		if err != nil {
			return nil, 0, errors.Annotate(err, "persist actions").Err()
		}

		out.updateWith(&ad)
		tally += len(batch)

		// TODO(b/248629691): A batch length of zero signals the successful end of the offload attempt.
		//                    Replace this with a better API for next.
		if len(batch) == 0 {
			return out, tally, nil
		}

		if err := insertBatch(ctx, a, batch); err != nil {
			return nil, 0, err
		}
	}
	return out, tally, nil
}

// persistObservations persists all of our observations associated with the actions found in `persistActions` to bigquery.
func persistObservations(ctx context.Context, a *actionRangePersistOptions, ad *ActionQueryAncillaryData) error {
	var hopper []*ObservationEntity
	query := datastore.NewQuery(ObservationKind).
		Gte("action_id", ad.SmallestID).
		Lte("action_id", ad.BiggestID)
	rErr := datastore.Run(ctx, query, func(o *ObservationEntity) error {
		hopper = append(hopper, o)
		if len(hopper) >= defaultBatchSize {
			if err := insertObservationBatch(ctx, a, hopper); err != nil {
				return errors.Annotate(err, "offloading records").Err()
			}
			hopper = nil
		}
		return nil
	})
	if rErr != nil {
		return errors.Annotate(rErr, "persisting observations").Err()
	}
	return insertObservationBatch(ctx, a, hopper)
}
