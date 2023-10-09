// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package misc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/models"
	"infra/cros/satlab/common/services/build_service"
	"infra/cros/satlab/common/site"
)

// StageAndWriteLocalStableVersion stages a recovery image to partner bucket and writes the associated rv metadata locally
func StageAndWriteLocalStableVersion(
	ctx context.Context,
	service build_service.IBuildService,
	rv *models.RecoveryVersion,
) error {
	buildVersion := strings.Split(rv.OsImage, "-")[1]
	bucket := site.GetGCSImageBucket()
	if bucket == "" {
		return errors.New("GCS_BUCKET not found")
	}
	_, err := service.StageBuild(ctx, rv.Board, rv.Model, buildVersion, bucket)
	if err != nil {
		return errors.Annotate(err, "stage stable version image to bucket").Err()
	}
	err = writeLocalStableVersion(rv, site.RecoveryVersionDirectory)
	if err != nil {
		return errors.Annotate(err, "write local stable version").Err()
	}
	return nil
}

// WriteLocalStableVersion saves a recovery version to the specified directory and creates the directory if necessary.
func writeLocalStableVersion(recovery_version *models.RecoveryVersion, path string) error {

	// Check if recovery_versions directory created
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	fname := fmt.Sprintf("%s%s-%s.json", path, recovery_version.Board, recovery_version.Model)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	// close file on exit and check for its returned error
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	rv, err := json.MarshalIndent(recovery_version, "", " ")
	if err != nil {
		return errors.Annotate(err, "marshal recovery version").Err()
	}
	_, err = f.Write(rv)
	if err != nil {
		return err
	}

	return nil
}
