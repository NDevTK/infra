// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cache"
	"infra/cros/recovery/internal/components/linux"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// cacheDownloadCheckExec performs download check by cache service.
func cacheDownloadCheckExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	testFilePath := argsMap.AsString(ctx, "test_path", "gs://cros-lab-servers/caching-backend/downloading-test.txt")
	log.Debugf(ctx, "Used file: %s", testFilePath)
	// Requesting convert GC path to caches service path.
	// Example: `http://Addr:8082/download/....`
	downloadPath, err := info.GetAccess().GetCacheUrl(ctx, info.GetDut().Name, testFilePath)
	if err != nil {
		return errors.Annotate(err, "cache download check").Err()
	}
	run := info.DefaultRunner()
	timeout := info.GetExecTimeout()
	curlArgs := []string{"-v"}

	// Update header to enforce download file always fresh.
	header := cache.HTTPRequestHeaders(ctx)
	header["X-NO-CACHE"] = "1"

	out, responseCode, err := linux.CurlURL(ctx, run, timeout, downloadPath, header, curlArgs...)
	log.Debugf(ctx, "Cache download output: %s", out)
	if err != nil {
		cache.RecordCacheAccessFailure(ctx, downloadPath, responseCode)
	}
	return errors.Annotate(err, "cache download check").Err()
}

func init() {
	execs.Register("cache_download_check", cacheDownloadCheckExec)
}
