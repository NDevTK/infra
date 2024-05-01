// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/golang/protobuf/jsonpb"

	build_api "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	lab_api "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros-tool-runner/internal/docker"
)

func startService(ctx context.Context, d *docker.Docker, block bool, netbind bool, service string) (*docker.Docker, error) {
	if err := d.Remove(ctx); err != nil {
		log.Printf("failed to clean up container %q. Error: %s", d.Name, err)
	}
	if err := d.Auth(ctx); err != nil {
		return d, errors.Annotate(err, "start service").Err()
	}
	if err := d.Run(ctx, block, netbind, service); err != nil {
		return d, errors.Annotate(err, "start service").Err()
	}
	return d, nil
}

// Create docker image path from ContainerImageInfo.
// Example or result: gcr.io/chromeos-bot/cros-provision:8835841547076258945
// Example of ContainerImageInfo:
//
//	"cros-provision" : {
//	  "repository" : { "hostname" : "gcr.io", "project" : "chromeos-bot" },
//	  "name"       : "cros-provision",
//	  "digest"     : "sha256:3e36d3622f5adad01080cc2120bb72c0714ecec6118eb9523586410b7435ae80",
//	  "tags" : [
//	    "8835841547076258945",
//	    "amd64-generic-release.R96-1.2.3"
//	  ]
//	}
func createImagePath(i *build_api.ContainerImageInfo) (string, error) {
	if i.GetName() == "" {
		return "", errors.Reason("create image path: name is empty").Err()
	}
	if i.GetRepository() == nil {
		return "", errors.Reason("create image path: no repository info").Err()
	}
	r := i.GetRepository()
	if r.GetHostname() == "" || r.GetProject() == "" {
		return "", errors.Reason("create image path: repository info is missing").Err()
	}
	if len(i.GetTags()) == 0 {
		return "", errors.Reason("create image path: no tags found").Err()
	}
	var path string
	if i.GetDigest() == "sha256:" {
		tag := i.GetTags()[0]
		path = fmt.Sprintf("%s/%s/%s:%s", r.GetHostname(), r.GetProject(), i.GetName(), tag)
		log.Println("Using tag as sha was not found.")
	} else {
		path = fmt.Sprintf("%s/%s/%s@%s", r.GetHostname(), r.GetProject(), i.GetName(), i.GetDigest())
	}
	log.Printf("container image: %q", path)
	return path, nil
}

// createRegistryName creates the Registry name used for authing to docker.
func createRegistryName(i *build_api.ContainerImageInfo) (string, error) {
	if i.GetRepository() == nil {
		return "", errors.Reason("create image path: no repository info").Err()
	}
	r := i.GetRepository()
	if r.GetHostname() == "" || r.GetProject() == "" {
		return "", errors.Reason("create image path: repository info is missing").Err()
	}
	return fmt.Sprintf("%s/%s", r.GetHostname(), r.GetProject()), nil
}

// createProvisionInput created input file for cros-provision.
func createProvisionInput(state *api.CrosProvisionRequest, dir string) error {
	if dir == "" {
		return errors.Reason("create input: directory is not provided").Err()
	}
	inputFilePath := path.Join(dir, InputFileName)
	f, err := os.Create(inputFilePath)
	if err != nil {
		return errors.Annotate(err, "create input").Err()
	}
	defer f.Close()

	marshaler := jsonpb.Marshaler{}
	if err := marshaler.Marshal(f, state); err != nil {
		return errors.Annotate(err, "create input").Err()
	}
	log.Printf("cros-provision request:" + state.String())

	err = f.Close()
	return errors.Annotate(err, "create input").Err()
}

func getAddr(i *lab_api.IpEndpoint) string {
	return fmt.Sprintf("%s:%d", i.GetAddress(), i.GetPort())
}
