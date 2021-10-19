// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clusteredfailures

import (
	"time"

	"cloud.google.com/go/bigquery"

	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/proto/google/descutil"

	bqpb "infra/appengine/weetbix/proto/bq"
	pb "infra/appengine/weetbix/proto/v1"

	"github.com/golang/protobuf/descriptor"
	desc "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

var tableMetadata *bigquery.TableMetadata

const partitionExpirationTime = 540 * 24 * time.Hour

const clusteredFailureMessage = "weetbix.bq.ClusteredFailure"

func init() {
	var err error
	var schema bigquery.Schema
	if schema, err = generateClusteredFailureSchema(); err != nil {
		panic(err)
	}

	tableMetadata = &bigquery.TableMetadata{
		TimePartitioning: &bigquery.TimePartitioning{
			Type:       bigquery.DayPartitioningType,
			Expiration: partitionExpirationTime,
			Field:      "partition_time",
		},
		Clustering: &bigquery.Clustering{
			Fields: []string{"cluster_algorithm", "cluster_id", "test_result_system", "test_result_id"},
		},
		// Relax ensures no fields are marked "required".
		Schema: schema.Relax(),
	}
}

func generateClusteredFailureSchema() (schema bigquery.Schema, err error) {
	fd, _ := descriptor.MessageDescriptorProto(&bqpb.ClusteredFailure{})
	// We also need to get FileDescriptorProto for StringPair, TestMetadata and FailureReason
	// because they are defined in different files.
	fdsp, _ := descriptor.MessageDescriptorProto(&pb.StringPair{})
	fdbtc, _ := descriptor.MessageDescriptorProto(&pb.BugTrackingComponent{})
	fdfr, _ := descriptor.MessageDescriptorProto(&pb.FailureReason{})
	fdprid, _ := descriptor.MessageDescriptorProto(&pb.PresubmitRunId{})
	fdset := &desc.FileDescriptorSet{File: []*desc.FileDescriptorProto{fd, fdsp, fdbtc, fdfr, fdprid}}
	return generateSchema(fdset, clusteredFailureMessage)
}

func generateSchema(fdset *desc.FileDescriptorSet, message string) (schema bigquery.Schema, err error) {
	conv := bq.SchemaConverter{
		Desc:           fdset,
		SourceCodeInfo: make(map[*desc.FileDescriptorProto]bq.SourceCodeInfoMap, len(fdset.File)),
	}
	for _, f := range fdset.File {
		conv.SourceCodeInfo[f], err = descutil.IndexSourceCodeInfo(f)
		if err != nil {
			return nil, errors.Annotate(err, "failed to index source code info in file %q", f.GetName()).Err()
		}
	}
	schema, _, err = conv.Schema(message)
	return schema, err
}
