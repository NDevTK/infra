// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clusteredfailures

import (
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"github.com/golang/protobuf/descriptor"
	desc "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/server/caching"
	"google.golang.org/protobuf/types/descriptorpb"

	"infra/appengine/weetbix/internal/bqutil"
	bqpb "infra/appengine/weetbix/proto/bq"
	pb "infra/appengine/weetbix/proto/v1"
)

// tableName is the name of the exported BigQuery table.
const tableName = "clustered_failures"

// schemaApplyer ensures BQ schema matches the row proto definitions.
var schemaApplyer = bq.NewSchemaApplyer(caching.RegisterLRUCache(50))

const partitionExpirationTime = 90 * 24 * time.Hour

const rowMessage = "weetbix.bq.ClusteredFailureRow"

var tableMetadata *bigquery.TableMetadata

// tableSchemaDescriptor is a self-contained DescriptorProto for describing
// row protocol buffers sent to the BigQuery Write API.
var tableSchemaDescriptor *descriptorpb.DescriptorProto

func init() {
	var err error
	var schema bigquery.Schema
	if schema, err = generateRowSchema(); err != nil {
		panic(err)
	}
	if tableSchemaDescriptor, err = generateRowSchemaDescriptor(); err != nil {
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

func generateRowSchema() (schema bigquery.Schema, err error) {
	fd, _ := descriptor.MessageDescriptorProto(&bqpb.ClusteredFailureRow{})
	// We also need to get FileDescriptorProto for StringPair, BugTrackingComponent, FailureReason
	// and PresubmitRunId because they are defined in different files.
	fdsp, _ := descriptor.MessageDescriptorProto(&pb.StringPair{})
	fdbtc, _ := descriptor.MessageDescriptorProto(&pb.BugTrackingComponent{})
	fdfr, _ := descriptor.MessageDescriptorProto(&pb.FailureReason{})
	fdprid, _ := descriptor.MessageDescriptorProto(&pb.PresubmitRunId{})
	fdcl, _ := descriptor.MessageDescriptorProto(&pb.Changelist{})
	fdset := &desc.FileDescriptorSet{File: []*desc.FileDescriptorProto{fd, fdsp, fdbtc, fdfr, fdprid, fdcl}}
	return bqutil.GenerateSchema(fdset, rowMessage)
}

func generateRowSchemaDescriptor() (*desc.DescriptorProto, error) {
	m := &bqpb.ClusteredFailureRow{}
	descriptorProto, err := adapt.NormalizeDescriptor(m.ProtoReflect().Descriptor())
	if err != nil {
		return nil, err
	}
	return descriptorProto, nil
}
