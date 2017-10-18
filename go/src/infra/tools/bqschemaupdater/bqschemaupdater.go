// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main takes as an argument the file path of a schema to be added to
// or updated in BigQuery. That file should contain a tabledef.TableDef proto
// in text format.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"go.chromium.org/luci/common/auth"
	"go.chromium.org/luci/common/proto"

	pb "infra/libs/bqschema/tabledef"
	"infra/libs/infraenv"
)

func loadProtoDescriptor(path string) (*descriptor.FileDescriptorSet, error) {
	descFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var desc descriptor.FileDescriptorSet
	err = protobuf.Unmarshal(descFileBytes, &desc)
	if err != nil {
		return nil, err
	}

	return &desc, nil
}

func protoDescToTableDef(fs *descriptor.FileDescriptorSet) (*pb.TableDef, error) {
	ret := &pb.TableDef{}
	for _, f := range fs.File {
		for _, mt := range f.MessageType {
			if mt.Options == nil {
				continue
			}
			log.Printf("checking MessageType: %q", *mt.Name)
			var datasetID, tableID, tableName, tableDescription *string
			var ok bool

			val, err := protobuf.GetExtension(mt.Options, pb.E_DatasetId)
			if err == nil {
				datasetID, ok = val.(*string)
				if !ok {
					log.Fatalf("Couldn't cast dataset_id to string.")
				}
			}

			val, err = protobuf.GetExtension(mt.Options, pb.E_TableId)
			if err == nil {
				tableID, ok = val.(*string)
				if !ok {
					log.Fatalf("Couldn't cast table_id to string.")
				}
			}

			val, err = protobuf.GetExtension(mt.Options, pb.E_TableName)
			if err == nil {
				tableName, ok = val.(*string)
				if !ok {
					log.Fatalf("Couldn't cast table_name to string.")
				}
			}

			val, err = protobuf.GetExtension(mt.Options, pb.E_TableDescription)
			if err == nil {
				tableDescription, ok = val.(*string)
				if !ok {
					log.Fatalf("Couldn't cast table_description to string.")
				}
			}

			if datasetID == nil && tableID != nil || datasetID != nil && tableID == nil {
				log.Fatalf("Must declare both dataset_id and table_id if either is declared.")
			}

			if datasetID != nil { // we have everything we need.
				log.Printf("Should create table: %q, %q, %q, %q", *datasetID, *tableID, *tableName, *tableDescription)

				ret.DatasetId = *datasetID
				ret.TableId = *tableID
				ret.Name = *tableName
				ret.Description = *tableDescription

				for _, f := range mt.Field {
					ret.Fields = append(ret.Fields, protoFieldToTabledefField(f))
				}
			} else {
				log.Printf("%q has options but not for bq schema.")
			}
		}
	}

	return ret, nil
}

func protoFieldToTabledefField(f *descriptor.FieldDescriptorProto) *pb.FieldSchema {
	return nil
}

func loadTableDef(path string) (*pb.TableDef, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	td := &pb.TableDef{}
	if err := proto.UnmarshalTextML(string(content), td); err != nil {
		return nil, err
	}
	return td, nil
}

func updateFromTableDef(ctx context.Context, ts tableStore, td *pb.TableDef) error {
	_, err := ts.getTableMetadata(ctx, td.DatasetId, td.TableId)
	if errNotFound(err) {
		var options []bigquery.CreateTableOption
		if td.PartitionTable {
			i := int(td.PartitionExpirationSeconds)
			tp := bigquery.TimePartitioning{Expiration: time.Duration(i) * time.Second}
			options = append(options, tp)
		}
		err = ts.createTable(ctx, td.DatasetId, td.TableId, options...)
		if err != nil {
			return err
		}
	}
	md := bigquery.TableMetadataToUpdate{
		Name:        td.Name,
		Description: td.Description,
		Schema:      pb.BQSchema(td.Fields),
	}
	err = ts.updateTable(ctx, td.DatasetId, td.TableId, md)
	return err
}

func main() {
	log.SetOutput(os.Stdout)
	ctx := context.Background()

	dry := flag.Bool("dry-run", false, "Only performs non-mutating operations; logs what would happen otherwise")
	project := flag.String("project", infraenv.ChromeInfraEventsProject, "Cloud project that the table belongs to.")
	useDescriptor := flag.Bool("use-descriptor", false, "Read input file as proto descriptor instead of text proto.")

	flag.Parse()
	file := flag.Arg(0)
	if file == "" {
		log.Fatal("Missing arg: file path for schema to add/update")
	}

	if *useDescriptor {
		desc, err := loadProtoDescriptor(file)
		log.Printf("desc, err: %v, %v", desc != nil, err)
		td, err := protoDescToTableDef(desc)
		log.Printf("def, err: %v, %v", td != nil, err)
	} // else {
	td, err := loadTableDef(file)
	if err != nil {
		log.Fatalf("Failed to load TableDef from %q: %s", file, err)
	}

	// Create an Authenticator and use it for BigQuery operations.
	authOpts := infraenv.DefaultAuthOptions()
	authOpts.Scopes = []string{bigquery.Scope}
	authenticator := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOpts)

	authTS, err := authenticator.TokenSource()
	if err != nil {
		log.Fatalf("Could not get authentication credentials: %s", err)
	}

	c, err := bigquery.NewClient(ctx, *project, option.WithTokenSource(authTS))
	if err != nil {
		log.Fatalf("Could not create BigQuery client: %s", err)
	}
	var ts tableStore
	ts = bqTableStore{c}
	if *dry {
		ts = dryRunTableStore{ts: ts, w: os.Stdout}
	}

	log.Printf("Updating dataset %q, table %q in project %q...", td.DatasetId, td.TableId, *project)
	err = updateFromTableDef(ctx, ts, td)
	if err != nil {
		log.Fatalf("Failed to update table: %s", err)
	}

	log.Println("Finished updating table.")
}
