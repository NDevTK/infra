// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package process3d

import (
	"testing"

	cloudPubsub "cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	buildPB "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
)

// TestProcessMessage_Success tests the ProcessMessage method with a successful release build.
func TestProcessMessage_Success(t *testing.T) {
	// Create a mock message with a successful release build report
	buildReport := &buildPB.BuildReport{
		Type: buildPB.BuildReport_BUILD_TYPE_RELEASE,
		Status: &buildPB.BuildReport_BuildStatus{
			Value: buildPB.BuildReport_BuildStatus_SUCCESS,
		},
	}
	data, err := proto.Marshal(buildReport)
	if err != nil {
		t.Fatal("error marshalling build report:", err)
	}
	msg := &cloudPubsub.Message{Data: data}

	p := NewProcess3d("", "", nil)

	err = p.ProcessMessage(msg)

	assert.NoError(t, err, "ProcessMessage should not return an error for a successful release build")
	assert.Len(t, p.buildMessages, 1, "ProcessMessage should add the message to buildMessages slice")
}

// TestProcessMessage_NonReleaseBuild tests the ProcessMessage method with a non-release build.
func TestProcessMessage_NonReleaseBuild(t *testing.T) {
	// Create a mock message with a non-release build report
	buildReport := &buildPB.BuildReport{
		Type: buildPB.BuildReport_BUILD_TYPE_FIRMWARE,
		Status: &buildPB.BuildReport_BuildStatus{
			Value: buildPB.BuildReport_BuildStatus_SUCCESS,
		},
	}
	data, err := proto.Marshal(buildReport)
	if err != nil {
		t.Fatal("error marshalling build report:", err)
	}
	msg := &cloudPubsub.Message{Data: data}

	p := NewProcess3d("", "", nil)

	err = p.ProcessMessage(msg)

	assert.NoError(t, err, "ProcessMessage should not return an error for a non-release build")
	assert.Empty(t, p.buildMessages, "ProcessMessage should not add the message to buildMessages slice for non-release builds")
}
