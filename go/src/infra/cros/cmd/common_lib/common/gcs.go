// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"go.chromium.org/chromiumos/config/go/build/api"
	"google.golang.org/protobuf/encoding/protojson"
)

// FetchImageData fetches container image metadata from provided gcs path
func FetchImageData(ctx context.Context, board string, gcsPath string) (map[string]*api.ContainerImageInfo, error) {
	gsutil := exec.CommandContext(ctx, "gsutil", "ls", "-l", gcsPath)
	sort := exec.CommandContext(ctx, "sort", "-k", "2")

	gPipe, err := gsutil.StdoutPipe()
	if err != nil {
		return nil, err
	}

	sort.Stdin = gPipe

	if err := gsutil.Start(); err != nil {
		return nil, err
	}
	imageDataRaw, err := sort.Output()
	if err != nil {
		return nil, err
	}

	regContainerEx := regexp.MustCompile(`gs://.*.jsonpb`)
	containerImages := regContainerEx.FindAllStringSubmatch(string(imageDataRaw), -1)

	if len(containerImages) == 0 {
		return nil, fmt.Errorf("Could not find any container images with given build %s", gcsPath)
	}
	archivePath := containerImages[len(containerImages)-1][0]

	cat := exec.CommandContext(ctx, "gsutil", "cat", archivePath)

	catOut, err := cat.Output()
	if err != nil {
		return nil, err
	}

	metadata := &api.ContainerMetadata{}
	err = protojson.Unmarshal(catOut, metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal metadata: %s", err)
	}

	images := metadata.Containers[board].Images

	Containers := make(map[string]*api.ContainerImageInfo)
	for _, image := range images {
		Containers[image.Name] = &api.ContainerImageInfo{
			Digest:     image.Digest,
			Repository: image.Repository,
			Name:       image.Name,
			Tags:       image.Tags,
		}
	}

	return Containers, nil
}