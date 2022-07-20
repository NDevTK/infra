// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package reclustering

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/tq"

	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/analysis/clusteredfailures"
	"infra/appengine/weetbix/internal/clustering/chunkstore"
	"infra/appengine/weetbix/internal/clustering/reclustering"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/tasks/taskspb"
)

const (
	taskClass = "reclustering"
	queue     = "reclustering"
)

var tc = tq.RegisterTaskClass(tq.TaskClass{
	ID:        taskClass,
	Prototype: &taskspb.ReclusterChunks{},
	Queue:     queue,
	Kind:      tq.NonTransactional,
})

// RegisterTaskHandler registers the handler for reclustering tasks.
func RegisterTaskHandler(srv *server.Server) error {
	ctx := srv.Context
	cfg, err := config.Get(ctx)
	if err != nil {
		return err
	}
	chunkStore, err := chunkstore.NewClient(ctx, cfg.ChunkGcsBucket)
	if err != nil {
		return err
	}
	srv.RegisterCleanup(func(context.Context) {
		chunkStore.Close()
	})

	cf, err := clusteredfailures.NewClient(ctx, srv.Options.CloudProject)
	if err != nil {
		return err
	}
	srv.RegisterCleanup(func(context.Context) {
		cf.Close()
	})

	analysis := analysis.NewClusteringHandler(cf)
	worker := reclustering.NewWorker(chunkStore, analysis)

	handler := func(ctx context.Context, payload proto.Message) error {
		task := payload.(*taskspb.ReclusterChunks)
		return reclusterTestResults(ctx, worker, task)
	}
	tc.AttachHandler(handler)
	return nil
}

// Schedule enqueues a task to recluster a range of chunks in a LUCI
// Project.
func Schedule(ctx context.Context, task *taskspb.ReclusterChunks) error {
	title := fmt.Sprintf("%s-%s-shard-%v", task.Project, task.AttemptTime.AsTime().Format("20060102-150405"), task.EndChunkId)
	return tq.AddTask(ctx, &tq.Task{
		Title: title,
		// Copy the task to avoid the caller retaining an alias to
		// the task proto passed to tq.AddTask.
		Payload: proto.Clone(task).(*taskspb.ReclusterChunks),
	})
}

func reclusterTestResults(ctx context.Context, worker *reclustering.Worker, task *taskspb.ReclusterChunks) error {
	next, err := worker.Do(ctx, task, reclustering.TargetTaskDuration)
	if err != nil {
		logging.Errorf(ctx, "Error re-clustering: %s", err)
		return err
	}
	if next != nil {
		if err := Schedule(ctx, next); err != nil {
			logging.Errorf(ctx, "Error scheduling continuation: %s", err)
			return err
		}
	}
	return nil
}
