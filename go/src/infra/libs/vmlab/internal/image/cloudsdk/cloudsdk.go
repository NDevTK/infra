// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cloudsdk

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"infra/libs/vmlab/api"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

const (
	project      = "betty-cloud-prototype"
	license      = "https://www.googleapis.com/compute/v1/projects/vm-options/global/licenses/enable-vmx"
	timeout      = 10 * time.Minute
	pollInterval = 30 * time.Second
)

// cloudsdkImageApi implementation api.ImageApi using Cloud SDK.
// go/cros-image-importer
type cloudsdkImageApi struct{}

// buildInfo contains information about a build.
type buildInfo struct {
	buildType    string
	board        string
	milestone    string
	majorVersion string
	minorVersion string
	patchNumber  string
	snapshot     string
	buildNumber  string
}

// New constructs a new api.ImageApi with CloudSDK backend.
func New() (api.ImageApi, error) {
	return &cloudsdkImageApi{}, nil
}

// computeImagesClient interfaces partial GCE client API
type computeImagesClient interface {
	Get(context.Context, *computepb.GetImageRequest, ...gax.CallOption) (*computepb.Image, error)
	Insert(context.Context, *computepb.InsertImageRequest, ...gax.CallOption) (*compute.Operation, error)
}

func (c *cloudsdkImageApi) GetImage(buildPath string, wait bool) (*api.GceImage, error) {
	buildInfo, err := parseBuildPath(buildPath)
	if err != nil {
		log.Default().Printf("Unable to parse build path: %v", err)
	}
	imageName := getImageName(buildInfo, buildPath)

	gceImage := &api.GceImage{
		Project: project,
		Name:    imageName,
		Source:  getGcsImagePath(buildPath),
	}

	var ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewImagesRESTClient: %v", err)
	}
	defer client.Close()

	op, gceImage, err := c.handle(client, buildInfo, gceImage)
	if err != nil {
		return nil, fmt.Errorf("Failed to import image: %w", err)
	}

	if wait {
		if op != nil {
			_ = op.Wait(ctx)
		} else {
			return c.poll(ctx, client, buildInfo, gceImage)
		}
	}
	return c.describeImage(client, gceImage)
}

// handle retrieves image info, and imports the image when not found.
func (c *cloudsdkImageApi) handle(client computeImagesClient, buildInfo *buildInfo, gceImage *api.GceImage) (*compute.Operation, *api.GceImage, error) {
	gceImage, _ = c.describeImage(client, gceImage)

	if gceImage.GetStatus() == api.GceImage_NOT_FOUND {
		op, err := c.importImage(client, buildInfo, gceImage)
		if err == nil {
			gceImage.Status = api.GceImage_PENDING
		}
		return op, gceImage, err
	}

	return nil, gceImage, nil
}

// poll retrieves image info periodically until error, success, or timeout.
func (c *cloudsdkImageApi) poll(ctx context.Context, client computeImagesClient, buildInfo *buildInfo, gceImage *api.GceImage) (*api.GceImage, error) {
	return gceImage, poll(ctx, func(ctx context.Context) (bool, error) {
		_, gceImage, err := c.handle(client, buildInfo, gceImage)
		if gceImage.GetStatus() == api.GceImage_READY {
			return true, err
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, pollInterval)
}

// describeImage detects the readiness of the given GCE image: READ, PENDING, or
// NOT_FOUND, which means import is needed.
func (c *cloudsdkImageApi) describeImage(client computeImagesClient, gceImage *api.GceImage) (*api.GceImage, error) {
	ctx := context.Background()

	// Create the gceImage
	req := &computepb.GetImageRequest{
		Image:   gceImage.Name,
		Project: gceImage.Project,
	}
	i, err := client.Get(ctx, req)

	gceImage.Status = api.GceImage_NOT_FOUND
	if i != nil {
		switch i.GetStatus() {
		case "READY":
			gceImage.Status = api.GceImage_READY
		case "PENDING":
			gceImage.Status = api.GceImage_PENDING
		}
	}

	return gceImage, err
}

// importImage imports an image from GCS into GCE.
func (c *cloudsdkImageApi) importImage(client computeImagesClient, buildInfo *buildInfo, image *api.GceImage) (*compute.Operation, error) {
	ctx := context.Background()

	// Create the image
	req := &computepb.InsertImageRequest{
		ImageResource: &computepb.Image{
			Licenses:    []string{license},
			Name:        &image.Name,
			Labels:      getImageLabels(buildInfo),
			Description: &image.Source,
			RawDisk: &computepb.RawDisk{
				Source: &image.Source,
			},
		},
		Project: image.Project,
	}

	return client.Insert(ctx, req)
}

// parseBuildPath parses build info from build path. Possible naming schemes:
// - betty-arc-r-release/R108-15178.0.0
// - betty-arc-r-cq/R108-15164.0.0-71927-8801111609984657185
// - betty-pi-arc-postsubmit/R113-15376.0.0-79071-8787141177342104481
func parseBuildPath(buildPath string) (*buildInfo, error) {
	re := regexp.MustCompile(`(.+)-(\w+)\/R(\d+)-(\d+).(\d+).(\d+)\-(\d+)\-(\d+)`)
	matches := re.FindStringSubmatch(buildPath)
	if len(matches) == 9 {
		return &buildInfo{
			buildType:    matches[2],
			board:        matches[1],
			milestone:    matches[3],
			majorVersion: matches[4],
			minorVersion: matches[5],
			patchNumber:  matches[6],
			snapshot:     matches[7],
			buildNumber:  matches[8],
		}, nil
	}
	re = regexp.MustCompile(`(.+)-(\w+)\/R(\d+)-(\d+).(\d+).(\d+)`)
	matches = re.FindStringSubmatch(buildPath)
	if len(matches) == 7 {
		return &buildInfo{
			buildType:    matches[2],
			board:        matches[1],
			milestone:    matches[3],
			majorVersion: matches[4],
			minorVersion: matches[5],
			patchNumber:  matches[6],
		}, nil
	}
	return nil, errors.New("Build path did not match known regex")
}

// getImageName generates an unique image name from buildInfo. If buildInfo is
// nil, generate name in format "unknown-{md5(buildPath)}" so that the same
// build always generates the same name, making image caching possible.
func getImageName(info *buildInfo, buildPath string) string {
	if info == nil {
		md5Hash := md5.Sum([]byte(buildPath))
		return "unknown-" + hex.EncodeToString(md5Hash[:])
	}

	imageName := strings.Join([]string{
		info.board, info.milestone, info.buildNumber, info.majorVersion,
		info.minorVersion, info.patchNumber, info.snapshot, info.buildType}, "-")

	// Image name have 63 characters limit.
	if len(imageName) > 63 {
		imageName = imageName[:63]
	}
	// Image name require lowercase letters.
	imageName = strings.ToLower(imageName)

	// Replace all remaining unsupported characters.
	re := regexp.MustCompile(`[^a-z0-9-]`)
	imageName = re.ReplaceAllString(imageName, "-")

	return imageName
}

// getImageLabels creates a map of labels for GCE image. For known images there
// are build-type, board, milestone. For unknown images build-type is set to
// unknown.
func getImageLabels(info *buildInfo) map[string]string {
	labels := make(map[string]string)
	if info != nil {
		labels["build-type"] = info.buildType
		labels["board"] = info.board
		labels["milestone"] = info.milestone
	} else {
		labels["build-type"] = "unknown"
	}
	return labels
}

// getGcsImagePath returns the full path of the tarball on GCS.
func getGcsImagePath(buildPath string) string {
	return fmt.Sprintf("https://storage.googleapis.com/chromeos-image-archive/%s/chromiumos_test_image_gce.tar.gz", buildPath)
}

// poll provides a generic implementation of calling f at interval, exits on
// error or ctx timeout. f return true to end poll early.
func poll(ctx context.Context, f func(context.Context) (bool, error), interval time.Duration) error {
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("context must have a deadline to avoid infinite polling")
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			success, err := f(ctx)
			if err != nil {
				return err
			}
			if success {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
