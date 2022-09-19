// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"go.chromium.org/luci/common/errors"
)

type FileEntry struct {
	Name     string
	Contents string
	Err      error
}

func readClose(reader io.ReadCloser, out []byte) (int, error) {
	if n, rErr := reader.Read(out); rErr != nil {
		if cErr := reader.Close(); cErr != nil {
			return 0, errors.Annotate(cErr, "error encountered while closing after failed read").Err()
		}
		return 0, rErr
	} else {
		if cErr := reader.Close(); cErr != nil {
			return 0, cErr
		}
		return n, nil
	}
}

func (e *extendedGSClient) ReadFile(ctx context.Context, bucket string, path string) ([]byte, error) {
	obj := e.Bucket(bucket).Object(path)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "read file").Err()
	}
	var out []byte
	if _, err := readClose(reader, out); err != nil {
		return nil, errors.Annotate(err, "read file").Err()
	}
	return out, nil
}

func (e *extendedGSClient) GetStatusLog(ctx context.Context, bucket string, date string, uuid string) (string, error) {
	prefix := fmt.Sprintf("test-runner/prod/%s/%s/", date, uuid)
	path := filepath.Join(prefix, "status.log")
	contents, err := e.ReadFile(ctx, bucket, path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
