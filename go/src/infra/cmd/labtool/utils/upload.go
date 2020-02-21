// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/gs"
)

const gsBucket = "cros-lab-inventory.appspot.com"
const scanLogPath = "assetScanLogs"

func upload(ctx context.Context, sc *storage.Client, localFilePath string, username string, bucketName string) error {
	ctx, abort := context.WithCancel(ctx)
	defer abort()
	bucket := sc.Bucket(bucketName)
	wr := bucket.Object(fmt.Sprintf("assetScanLogs/%s/%s", username, filepath.Base(localFilePath))).NewWriter(ctx)
	logReader, err := os.Open(localFilePath)
	defer logReader.Close()
	if err != nil {
		return err

	}
	if _, err := io.Copy(wr, logReader); err != nil {
		return errors.Annotate(err, "upload %s to %s", localFilePath, bucketName).Err()
	}
	if err := wr.Close(); err != nil {
		return errors.Annotate(err, "failed to finalize the upload").Err()
	}
	return nil
}

func upload2(sc gs.Client, localFilePath string, username string) error {
	p := gs.Path(fmt.Sprintf("gs://%s/%s/%s/%s", gsBucket, scanLogPath, username, filepath.Base(localFilePath)))
	wr, err := sc.NewWriter(p)
	if err != nil {
		return err
	}
	logReader, err := os.Open(localFilePath)
	defer logReader.Close()
	if err != nil {
		return err

	}
	if _, err := io.Copy(wr, logReader); err != nil {
		return errors.Annotate(err, "upload %s to %s", localFilePath, gsBucket).Err()
	}
	if err := wr.Close(); err != nil {
		return errors.Annotate(err, "failed to finalize the upload").Err()
	}
	return nil
}
