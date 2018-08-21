// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chromiumbuildstats

import (
	"context"
	"infra/appengine/chromium_build_stats/ninjalog"

	"cloud.google.com/go/bigquery"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/bq"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	bqDataset     = "ninjalog"
	bqResultTable = "staging"
)

func sendBigquery(ctx context.Context, info ninjalog.NinjaLog) error {
	ninjaTasks, err := ninjalog.ConvertNinjalog(info)
	if err != nil {
		log.Errorf(ctx, "failed to convert ninjalog: %v", err)
		return err
	}

	projectID := appengine.AppID(ctx)
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Errorf(ctx, "failed to create new client: %v", err)
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Warningf(ctx, "failed to close client: %v", err)
		}
	}()

	up := bq.NewUploader(ctx, client, bqDataset, bqResultTable)
	m := make([]proto.Message, len(ninjaTasks))
	for i, t := range ninjaTasks {
		m[i] = t
	}

	if err := up.Put(ctx, m...); err != nil {
		log.Errorf(ctx, "failed to put BigQuery: %v", err)
		return err
	}
	log.Debugf(ctx, "success to send bigquery!")

	return nil
}
