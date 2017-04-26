// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	swarmbucket "github.com/luci/luci-go/common/api/buildbucket/swarmbucket/v1"
	swarming "github.com/luci/luci-go/common/api/swarming/swarming/v1"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/errors"
)

func grabBuilderDefinition(ctx context.Context, bbHost, bucket, builder string, authOpts auth.Options) (*JobDefinition, error) {
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)
	authClient, err := authenticator.Client()
	if err != nil {
		return nil, err
	}
	sbucket, err := swarmbucket.New(authClient)
	sbucket.BasePath = fmt.Sprintf("%s/api/swarmbucket/v1/", bbHost)

	type parameters struct {
		BuilderName string `json:"builder_name"`
	}

	data, err := json.Marshal(&parameters{builder})
	if err != nil {
		return nil, err
	}

	args := &swarmbucket.SwarmingSwarmbucketApiGetTaskDefinitionRequestMessage{
		BuildRequest: &swarmbucket.ApiPutRequestMessage{
			Bucket:         bucket,
			ParametersJson: string(data),
		},
	}
	answer, err := sbucket.GetTaskDef(args).Context(ctx).Do()
	if err != nil {
		return nil, errors.WrapTransient(err)
	}

	newTask := &swarming.SwarmingRpcsNewTaskRequest{}
	r := strings.NewReader(answer.TaskDefinition)
	if err := json.NewDecoder(r).Decode(newTask); err != nil {
		return nil, err
	}

	return JobDefinitionFromNewTaskRequest(newTask)
}
