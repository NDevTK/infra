// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package process3d handles collecting all pubsub messages to determine readiness
// and execution of 3d configs
package process3d

import (
	"context"
	"fmt"
	"sync"

	cloudPubsub "cloud.google.com/go/pubsub"

	buildPB "go.chromium.org/chromiumos/infra/proto/go/chromiumos"

	"infra/cros/cmd/kron/common"
	"infra/cros/cmd/kron/configparser"
	"infra/cros/cmd/kron/pubsub"
)

// Process3d is the struct for processing 3d configs
type Process3d struct {
	buildMessages []*cloudPubsub.Message
	buildInfoMap  map[int64][]*string
	psClient      pubsub.ReceiveClient
	mutex         sync.Mutex
	projectID     string
	subscription  string
	configs       configparser.ConfigList
}

// NewProcess3d creates a new instance of Process3d struct
func NewProcess3d(projectID string, subscription string, config configparser.ConfigList) *Process3d {
	return &Process3d{
		psClient:     nil,
		buildInfoMap: make(map[int64][]*string), // Initialize the map
		projectID:    projectID,
		subscription: subscription,
		configs:      config,
	}
}

// AddToBuildInfoMap appends build target to buildId map
func (p *Process3d) AddToBuildInfoMap(key int64, value string) {

	if list, ok := p.buildInfoMap[key]; ok {
		// Key exists, update the list
		p.buildInfoMap[key] = append(list, &value)
	} else {
		// Key doesn't exist, create a new list with the string
		p.buildInfoMap[key] = []*string{&value}
	}
}

// ProcessMessage stores each messages in a map. It does not ack/Nack messages. It's acked accordingly in finalize.
// Note: If messages are not acked in "finalize", receiver client will be blocking even if context is cancelled.
func (p *Process3d) ProcessMessage(msg *cloudPubsub.Message) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	buildReport := buildPB.BuildReport{}
	err := common.ProtoUnmarshaller.Unmarshal(msg.Data, &buildReport)
	if err != nil {
		return err
	}
	// Check for a successful release build. Ignore all types of reports.
	if !(buildReport.Type == buildPB.BuildReport_BUILD_TYPE_RELEASE && buildReport.Status.Value.String() == "SUCCESS") {
		msg.Ack()
		return nil
	}
	p.buildMessages = append(p.buildMessages, msg)
	p.AddToBuildInfoMap(buildReport.GetParent().GetBuildbucketId(), buildReport.GetConfig().GetTarget().String())

	return nil
}

// Process3d starts a pubsub client to receive messages. It collects it, checks if 3d is ready to be triggered.
// If ready, it creates ctp request for configs and then makes bb request. It sends "finalize" method
// to evaluate if 3d is ready and process accordingly.
func (p *Process3d) Process3d() error {
	common.Stdout.Println("Initializing Pub/Sub Receive Client")
	ctx := context.Background()
	psClient, err := pubsub.InitReceiveClientWithTimer(ctx, p.projectID, p.subscription, p.ProcessMessage)
	if err != nil {
		common.Stdout.Println("error initializing pubsub receive client")
		return err
	}

	common.Stdout.Println("Pulling messages from Pub/Sub Queue")
	err = psClient.PullAllMessagesForProcessing(p.finalize)
	if err != nil {
		return err
	}

	return nil
}

func (p *Process3d) finalize() {
	fmt.Println(len(p.buildMessages))
	for id, buildTargetList := range p.buildInfoMap {
		fmt.Printf("Buildbicket id %d , number of builds : %d", id, len(buildTargetList))
		fmt.Println()
	}
	// TODO Determine which parent buildId is to be used.
	// TODO Check if parent buildId is complete.
	// TODO Create CTP reqs for 3d config.
	// TODO Make a BB reqs for those CTP reqs.

	// Nack all the message for now to free pubsub "receive" blocking call
	for _, msg := range p.buildMessages {
		msg.Nack()
	}

}
