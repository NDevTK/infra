// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ingestion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms"
	cpb "infra/appengine/weetbix/internal/clustering/proto"
	"infra/appengine/weetbix/internal/clustering/state"
	pb "infra/appengine/weetbix/proto/v1"

	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server/span"
)

// Options represents parameters to the ingestion.
type Options struct {
	// Project is the LUCI Project.
	Project string
	// PartitionTime is the start of the retention period of test results
	// being ingested.
	PartitionTime time.Time
	// Realm is the LUCI Realm of the test results.
	Realm string
	// InvocationID is the identity of the invocation being ingested.
	InvocationID string
	// PresubmitRunID is the identity of the presubmit run (if any).
	PresubmitRunID *pb.PresubmitRunId
}

// Analysis is the interface for cluster analysis.
type Analysis interface {
	// ClustersUpdated handles (re-)clustered test results. It is called after
	// the spanner transaction effecting the (re-)clustering has committed.
	// commitTime is the Spanner time the transaction committed.
	ClustersUpdated(ctx context.Context, updates *clustering.Update, commitTime time.Time) error
}

// ChunkStore is the interface for the blob store archiving chunks of test
// results for later re-clustering.
type ChunkStore interface {
	// Put saves the given chunk to storage. If successful, it returns
	// the randomly-assigned ID of the created object.
	Put(ctx context.Context, project string, content *cpb.Chunk) (string, error)
}

// ChunkSize is the number of test failures that are to appear in each chunk.
const ChunkSize = 1000

// Ingester handles the ingestion of test results for clustering.
type Ingester struct {
	chunkStore ChunkStore
	analysis   Analysis
}

// New initialises a new Ingester.
func New(cs ChunkStore, a Analysis) *Ingester {
	return &Ingester{
		chunkStore: cs,
		analysis:   a,
	}
}

// Ingest ingests test results for clustering.
func (i *Ingester) Ingest(ctx context.Context, opts Options, tvs []*rdbpb.TestVariant) error {
	failures := failuresFromTestVariants(opts, tvs)
	project := opts.Project

	chunks := chunk(failures)

	// For each chunk:
	// - Archive the test results.
	// - Cluster the failures.
	// - Write out the chunk state.
	// - Perform analysis.
	for p, chunk := range chunks {
		// Derive a chunkID deterministically from the ingested root invocation
		// ID and page number. In case of retry this avoids ingesting the same
		// data twice.
		id := chunkID(opts.InvocationID, p)

		readCtx := span.Single(ctx)
		_, err := state.Read(readCtx, project, id)
		if err != state.NotFound {
			if err == nil {
				// Chunk was already ingested as part of an earlier ingestion attempt.
				// Do not attempt to ingest again.
				continue
			}
			return err
		}

		// Upload the chunk. The objectID is randomly generated each time
		// so the actual insertion of the chunk will be atomic with the
		// ClusteringState row in Spanner.
		objectID, err := i.chunkStore.Put(ctx, project, chunk)
		if err != nil {
			return err
		}

		clusterResults := algorithms.Cluster(chunk.Failures)

		clusterState := &state.Entry{
			Project:           project,
			ChunkID:           id,
			PartitionTime:     opts.PartitionTime,
			ObjectID:          objectID,
			AlgorithmsVersion: algorithms.AlgorithmsVersion,
			RuleVersion:       clusterResults.RuleVersion,
			Clusters:          clusterResults.Clusters,
		}
		f := func(ctx context.Context) error {
			// Could fail due to data race if ingestion for this chunk was already
			// completed by a duplicate thread. This is OK.
			if err := state.Create(ctx, clusterState); err != nil {
				return err
			}
			return nil
		}
		commitTime, err := span.ReadWriteTransaction(ctx, f)
		if err != nil {
			return err
		}

		update := &clustering.Update{
			Project: project,
			ChunkID: id,
			Updates: prepareClusterUpdates(chunk, clusterResults),
		}
		if err := i.analysis.ClustersUpdated(ctx, update, commitTime); err != nil {
			return err
		}
	}
	return nil
}

func prepareClusterUpdates(chunk *cpb.Chunk, clusterResults *algorithms.ClusterResults) []*clustering.FailureUpdate {
	var updates []*clustering.FailureUpdate
	for i, testResult := range chunk.Failures {
		update := &clustering.FailureUpdate{
			TestResult:  testResult,
			NewClusters: clusterResults.Clusters[i],
		}
		updates = append(updates, update)
	}
	return updates
}

// chunk determinisitically divides the specified failures into chunks.
func chunk(failures []*cpb.Failure) []*cpb.Chunk {
	var result []*cpb.Chunk
	pages := (len(failures) + (ChunkSize - 1)) / ChunkSize
	for p := 0; p < pages; p++ {
		start := p * ChunkSize
		end := start + ChunkSize
		if end > len(failures) {
			end = len(failures)
		}
		page := failures[start:end]

		for i, f := range page {
			f.ChunkIndex = int64(i)
		}

		c := &cpb.Chunk{
			Failures: page,
		}
		result = append(result, c)
	}
	return result
}

// chunkID generates an identifier for the chunk deterministically.
// The identifier will be 32 lowercase hexadecimal characters. Generated
// identifiers will be approximately evenly distributed through
// the keyspace.
func chunkID(rootInvocationID string, seq int) string {
	content := fmt.Sprintf("%q:%v", rootInvocationID, seq)
	sha256 := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sha256[:16])
}
