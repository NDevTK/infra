// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package subcommands

import (
	"context"

	"github.com/maruel/subcommands"
	"google.golang.org/protobuf/encoding/protojson"

	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/firestore"
)

// firestoreCommand is the interface for the firestore-sync subcommand
type firestoreCommand struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	isProd bool
}

// setFlags creates the flags that the user can set.
func (c *firestoreCommand) setFlags() {
	c.Flags.BoolVar(&c.isProd, "prod", false, "If provided the production project environment will be used.")
}

// GetFirestoreCommand creates the subcommand with CLI flags.
func GetFirestoreCommand(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "firestore-sync <options>",
		LongDesc:  "Sync the firestore DB with the current ToT SuiteScheduler.cfg.",
		ShortDesc: "Sync the firestore DB with the current ToT SuiteScheduler.cfg.",
		CommandRun: func() subcommands.CommandRun {
			cmd := &firestoreCommand{}
			cmd.authFlags = authcli.Flags{}
			cmd.authFlags.Register(cmd.GetFlags(), authOpts)
			cmd.setFlags()
			return cmd
		},
	}
}

// fetchToTConfigs fetches the ToT configs and returns the config list.
func fetchToTConfigs() ([]*suschpb.SchedulerConfig, error) {
	// Ingest the ToT Configs
	configBytes, err := common.FetchFileFromURL(common.SuiteSchedulerCfgURL)
	if err != nil {
		return nil, err
	}

	// Convert the bytes into a proto object
	configs, err := configparser.BytesToSchedulerProto(configBytes)
	if err != nil {
		return nil, err
	}
	return configs.Configs, nil
}

// generateFirestoreItemList returns a properly formatted Firestore item for us
// to insert into the database.
func generateFirestoreItemList(configs []*suschpb.SchedulerConfig) ([]*firestore.FirestoreItem, error) {
	insertItems := []*firestore.FirestoreItem{}
	for _, config := range configs {
		jsonData, err := protojson.Marshal(config)
		if err != nil {
			return nil, err
		}

		datum := map[string]string{
			"configJSON": string(jsonData),
		}

		firestoreItem := firestore.FirestoreItem{
			DocName: config.Name,
			Datum:   datum,
		}
		insertItems = append(insertItems, &firestoreItem)
	}
	return insertItems, nil
}

// Run is the "main()" of the firestore sync command.
func (c *firestoreCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := context.Background()

	// Get the target ProjectID.
	projectID := common.StagingProjectID
	if c.isProd {
		projectID = common.ProdProjectID
	}

	// Initialize the client at the target projectID.
	firestoreClient, err := firestore.InitClient(ctx, projectID, common.FirestoreDatabaseName)
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}

	// Form the collectionRef for the configs Collection in Firestore.
	//
	// NOTE: If the name or structure changes this will ned to be updated.
	configCollection := firestoreClient.Collection(common.FirestoreConfigCollectionName)

	configs, err := fetchToTConfigs()
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}

	// Generate the list of items to send to the firestore client.
	insertItems, err := generateFirestoreItemList(configs)
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}

	// Batch write the config items to firestore.
	writeJobResults, err := firestore.BatchSet(ctx, configCollection, firestoreClient, insertItems)
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}

	// Handle result errors
	for _, job := range writeJobResults {
		if _, err := job.Results(); err != nil {
			common.Stderr.Println(err)
			return 1

		}
	}

	return 0
}
