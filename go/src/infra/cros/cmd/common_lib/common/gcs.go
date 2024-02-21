// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

var ErrBucketNotExist = errors.New("bucket does not exist")
var ErrObjectNotExist = errors.New("object does not exist")

// actionTimeout is the timeout that pertains to reading or writing
// a single file from/to GCS.
const actionTimeout = 60 * time.Second

// Storage metadata for remote file
type GSObject struct {
	Bucket string
	Object string
}

// cleanObjectPath removes the first slash in a path if any. The built-in path
// framework concerns itself with removing trailing only, and as such we do so
// bespoke.
func cleanObjectPath(objectPath string) string {
	if strings.HasPrefix(objectPath, "/") {
		return objectPath[1:]
	}
	return objectPath
}

func ParseGSURL(gsURL string) (GSObject, error) {
	if !strings.HasPrefix(gsURL, "gs://") {
		return GSObject{}, fmt.Errorf("gs url must begin with 'gs://', instead have, %s", gsURL)
	}

	u, err := url.Parse(gsURL)
	if err != nil {
		return GSObject{}, fmt.Errorf("unable to parse url, %w", err)
	}

	// Host corresponds to bucket
	// Path corresponds to object (though we need to remove prepending '/')
	return GSObject{
		Bucket: u.Host,
		Object: cleanObjectPath(u.Path),
	}, nil
}

func GetURLPath(gsURL string) (string, error) {
	if !strings.HasPrefix(gsURL, "gs://") {
		return "", fmt.Errorf("gs url must begin with 'gs://', instead have, %s", gsURL)
	}

	return gsURL[len("gs://"):], nil

}

// DownloadFile downloads a file from a designated gsURL to a given
// path on the local file system. If the bucket does not exist,
// returns ErrBucketNotExist. If the object does not exist,
// returns ErrObjectNotExist.
func DownloadFile(ctx context.Context, client *storage.Client, gsURL, destLocalPath string) error {
	logging.Infof(ctx, "Starting download of %s to %s", gsURL, destLocalPath)
	object, err := ParseGSURL(gsURL)
	if err != nil {
		return fmt.Errorf("unable to parse gs url, %w", err)
	}
	logging.Infof(ctx, "Parsed URL: %s", object)
	err = Read(ctx, client, object, destLocalPath)
	if errors.Is(err, ErrBucketNotExist) {
		return ErrBucketNotExist
	}
	if err == ErrObjectNotExist {
		return ErrObjectNotExist
	}
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	return nil
}

func NewStorageClientWithDefaultAccount(ctx context.Context) (*storage.Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Read downloads a file from GCS to the given local path. If the bucket does
// not exist in GCS, the method returns ErrBucketNotExist. If the object does
// not exist in GCS, the method returns ErrObjectNotExist.
func Read(ctx context.Context, client *storage.Client, gsObject GSObject, destFilePath string) (retErr error) {
	ctx, cancel := context.WithTimeout(ctx, actionTimeout)
	defer cancel()

	reader, err := client.Bucket(gsObject.Bucket).Object(gsObject.Object).NewReader(ctx)
	if errors.Is(err, storage.ErrBucketNotExist) {
		return ErrBucketNotExist
	}
	if errors.Is(err, storage.ErrObjectNotExist) {
		return ErrObjectNotExist
	}
	if err != nil {
		return fmt.Errorf("opening reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	dst, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening destination file %q: %w", destFilePath, err)
	}
	defer func() {
		if err := dst.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("copying to local file: %w", err)
	}
	return nil
}

// FetchImageData fetches container image metadata from provided gcs path
func FetchImageData(ctx context.Context, board string, gcsPath string) (map[string]*api.ContainerImageInfo, error) {
	if !strings.HasSuffix(gcsPath, ContainerMetadataPath) {
		gcsPath = gcsPath + ContainerMetadataPath
		logging.Infof(ctx, "container metadata path created: %s", gcsPath)
	}

	cat := exec.CommandContext(ctx, "gsutil", "cat", gcsPath)

	catOut, err := cat.Output()
	if err != nil {
		logging.Infof(ctx, "error while downloading container metadata: %s", err)
		return nil, errors.Annotate(err, "error while downloading container metadata: ").Err()
	}

	metadata := &api.ContainerMetadata{}
	err = protojson.Unmarshal(catOut, metadata)
	if err != nil {
		logging.Infof(ctx, "error while downloading metadata: %s", err)
		return nil, errors.Annotate(err, "unable to unmarshal metadata: ").Err()
	}

	imagesMap, ok := metadata.Containers[board]
	if !ok {
		logging.Infof(ctx, fmt.Sprintf("provided board %s not found in container images map", board))
		return nil, fmt.Errorf("provided board %s not found in container images map", board)
	}

	Containers := make(map[string]*api.ContainerImageInfo)
	for _, image := range imagesMap.GetImages() {
		Containers[image.Name] = &api.ContainerImageInfo{
			Digest:     image.Digest,
			Repository: image.Repository,
			Name:       image.Name,
			Tags:       image.Tags,
		}
	}

	return Containers, nil
}

// DownloadGcsFileToLocal downloads gcs file to local if it doesn't exist.
func DownloadGcsFileToLocal(ctx context.Context, gcsPath string, tempRootDir string) (string, error) {
	client, err := NewStorageClientWithDefaultAccount(ctx)
	if err != nil {
		logging.Infof(ctx, "error while creating new storage client: %s", err)
		return "", err
	}

	urlPath, err := GetURLPath(gcsPath)
	if err != nil {
		logging.Infof(ctx, "error while gettting url path: %s", err)
		return "", err
	}
	localFilePath := path.Join(tempRootDir, urlPath)
	err = os.MkdirAll(filepath.Dir(localFilePath), os.ModePerm)
	if err != nil {
		logging.Infof(ctx, "error while making local dir: %s", err)
		return "", err
	}

	logging.Infof(ctx, "local gcs file path: %s", localFilePath)
	if CheckIfFileExists(localFilePath) != nil {
		err = DownloadFile(ctx, client, gcsPath, localFilePath)
		if err != nil {
			logging.Infof(ctx, "error while downloading file: %s", err)
			return "", err
		}
	} else {
		logging.Infof(ctx, "local gcs file already exists")
	}

	return localFilePath, nil
}
