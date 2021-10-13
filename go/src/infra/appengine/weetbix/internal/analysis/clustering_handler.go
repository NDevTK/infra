package analysis

import (
	"context"
	"encoding/hex"
	"time"

	"infra/appengine/weetbix/internal/analysis/clusteredfailures"
	"infra/appengine/weetbix/internal/clustering"
	cpb "infra/appengine/weetbix/internal/clustering/proto"

	"cloud.google.com/go/bigquery"
)

// ClusteringHandler handles test result (re-)clustering events, to
// ensure analysis remains up-to-date.
type ClusteringHandler struct {
	clusteredFailures ClusteredFailuresClient
}

// ClusteredFailuresClient exports clustered failures to BigQuery for
// further analysis.
type ClusteredFailuresClient interface {
	// Insert inserts the given rows in BigQuery.
	Insert(ctx context.Context, rows []*clusteredfailures.Entry) error
}

func NewClusteringHandler(cf ClusteredFailuresClient) *ClusteringHandler {
	return &ClusteringHandler{
		clusteredFailures: cf,
	}
}

// ClustersUpdated handles (re-)clustered test results. It is called after
// the spanner transaction effecting the (re-)clustering has committed.
// commitTime is the Spanner time the transaction committed.
//
// If this method fails, it will not be retried and data loss or inconsistency
// (in this method's BigQuery export) may occur. This could be improved in
// future with a two-stage apply process (journalling the BigQuery updates
// to be applied as part of the original transaction and retrying them at
// a later point if they do not succeed).
func (r *ClusteringHandler) ClustersUpdated(ctx context.Context, updates *clustering.Update, commitTime time.Time) error {
	rowUpdates := prepareDelta(updates, commitTime)
	return r.clusteredFailures.Insert(ctx, rowUpdates)
}

// prepareDelta prepares entries into the BigQuery clustered failures table in
// response to a reclustering.
func prepareDelta(updates *clustering.Update, commitTime time.Time) []*clusteredfailures.Entry {
	var result []*clusteredfailures.Entry
	for _, u := range updates.Updates {
		deleted := make(map[string]*clustering.ClusterRef)
		retained := make(map[string]*clustering.ClusterRef)
		new := make(map[string]*clustering.ClusterRef)

		previousHasBugCluster := false
		for _, pc := range u.PreviousClusters {
			deleted[pc.Key()] = pc
			if isBugCluster(pc) {
				previousHasBugCluster = true
			}
		}
		newHasBugCluster := false
		for _, nc := range u.NewClusters {
			key := nc.Key()
			if _, ok := deleted[key]; ok {
				delete(deleted, key)
				retained[key] = nc
			} else {
				new[key] = nc
			}
			if isBugCluster(nc) {
				newHasBugCluster = true
			}
		}
		// Create rows for deletions.
		for _, dc := range deleted {
			isIncluded := false
			isIncludedWithHighPriority := false
			row := entryFromUpdate(updates.Project, updates.ChunkID, dc, u.TestResult, isIncluded, isIncludedWithHighPriority, commitTime)
			result = append(result, row)
		}
		// Create rows for retained clusters for which inclusion was modified.
		for _, rc := range retained {
			isIncluded := true
			previousIncludedWithHighPriority := !previousHasBugCluster || isBugCluster(rc)
			newIncludedWithHighPriority := !newHasBugCluster || isBugCluster(rc)
			if previousIncludedWithHighPriority == newIncludedWithHighPriority {
				// The inclusion status of the test result in the cluster has not changed.
				// For efficiency, do not stream an update.
				continue
			}
			row := entryFromUpdate(updates.Project, updates.ChunkID, rc, u.TestResult, isIncluded, newIncludedWithHighPriority, commitTime)
			result = append(result, row)
		}
		// Create rows for new clusters.
		for _, nc := range new {
			isIncluded := true
			isIncludedWithHighPriority := !newHasBugCluster || isBugCluster(nc)
			row := entryFromUpdate(updates.Project, updates.ChunkID, nc, u.TestResult, isIncluded, isIncludedWithHighPriority, commitTime)
			result = append(result, row)
		}
	}
	return result
}

func isBugCluster(c *clustering.ClusterRef) bool {
	// TODO(crbug.com/1243174): When failure association rules are implemented,
	// return whether the clustering algorithm is the failure association rule
	// clustering algorithm.
	return false
}

func entryFromUpdate(project, chunkID string, cluster *clustering.ClusterRef, failure *cpb.Failure, included, includedWithHighPriority bool, commitTime time.Time) *clusteredfailures.Entry {
	entry := &clusteredfailures.Entry{
		Project:          project,
		ClusterAlgorithm: cluster.Algorithm,
		ClusterID:        hex.EncodeToString(cluster.ID),
		TestResultID:     failure.TestResultId,
		LastUpdated:      commitTime,

		PartitionTime: failure.PartitionTime.AsTime(),

		IsIncluded:                 included,
		IsIncludedWithHighPriority: includedWithHighPriority,

		ChunkID:    chunkID,
		ChunkIndex: failure.ChunkIndex,

		Realm:                     failure.Realm,
		TestID:                    failure.TestId,
		Variant:                   variantFromProto(failure.Variant),
		VariantHash:               failure.VariantHash,
		FailureReason:             failureReasonFromProto(failure.FailureReason),
		Component:                 failure.Component,
		StartTime:                 failure.StartTime.AsTime(),
		Duration:                  failure.Duration.AsDuration(),
		IsExonerated:              failure.IsExonerated,
		RootInvocationID:          failure.RootInvocationId,
		RootInvocationResultSeq:   failure.RootInvocationResultSeq,
		RootInvocationResultCount: failure.RootInvocationResultCount,
		IsRootInvocationBlocked:   failure.IsRootInvocationBlocked,
		TaskID:                    failure.TaskId,
		TaskResultSeq:             failure.TaskResultSeq,
		TaskResultCount:           failure.TaskResultCount,
		IsTaskBlocked:             failure.IsTaskBlocked,
		CQRunID:                   bigquery.NullString{},
	}
	if failure.CqId != "" {
		entry.CQRunID = bigquery.NullString{Valid: true, StringVal: failure.CqId}
	}
	return entry
}

func variantFromProto(v *cpb.Variant) []*clusteredfailures.Variant {
	var result []*clusteredfailures.Variant
	for k, v := range v.Def {
		result = append(result, &clusteredfailures.Variant{
			Key:   k,
			Value: v,
		})
	}
	return result
}

func failureReasonFromProto(fr *cpb.FailureReason) *clusteredfailures.FailureReason {
	if fr == nil {
		return nil
	}
	return &clusteredfailures.FailureReason{
		PrimaryErrorMessage: fr.PrimaryErrorMessage,
	}
}
