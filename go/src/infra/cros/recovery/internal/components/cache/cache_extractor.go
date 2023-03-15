// Copyright (c) 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
)

// ExtractRequest holds all data required to extract file from file on cache service.
type ExtractRequest struct {
	// URL to download the file from cache service.
	CacheFileURL string
	// Name of the file we wantt o extract from file.
	ExtractFileName string
	// Filepath of destination file.
	DestintionFilePath string
	// Download timeout.
	Timeout time.Duration
	// Number of times the download can be re-attempted.
	DownloadImageReattemptCount int
	// Wait before the download is re-attempted.
	DownloadImageReattemptWait time.Duration
}

// Extract extract file from cache service by modifying URL to download the file.
func Extract(ctx context.Context, req *ExtractRequest, run components.Runner) error {
	// Path provided by TLS cannot be used for downloading and/or extracting the image file.
	// But we can utilize the address of caching service and apply some string manipulation to construct the URL that can be used for this.
	// Example: `http://Addr:8082/extract/chromeos-image-archive/board-release/R99-XXXXX.XX.0/chromiumos_test_image.tar.xz?file=chromiumos_test_image.bin`
	extractPath := strings.Replace(req.CacheFileURL, "/download/", "/extract/", 1)
	sourcePath := fmt.Sprintf("%s?file=%s", extractPath, req.ExtractFileName)
	// We need to count the original download as well as any re-attempts.
	remainingDownloadAttempts := req.DownloadImageReattemptCount + 1
	for {
		if httpResponseCode, err := CurlFile(ctx, run, sourcePath, req.DestintionFilePath, req.Timeout); err != nil {
			log.Debugf(ctx, "Extract: HTTP Response Code is :%d", httpResponseCode)
			if httpResponseCode/100 == 5 && remainingDownloadAttempts > 1 {
				remainingDownloadAttempts -= 1
				time.Sleep(req.DownloadImageReattemptWait)
				continue
			}
			return errors.Annotate(err, "extract from cache").Err()
		} else {
			break
		}
	}
	return nil
}

// CurlFile downloads file by using curl util.
func CurlFile(ctx context.Context, run components.Runner, sourcePath, destinationPath string, timeout time.Duration) (int, error) {
	out, err := run(ctx, timeout, "curl", sourcePath, "--output", destinationPath, "--fail")
	if err == nil {
		log.Debugf(ctx, "Successfully download %q from %q", destinationPath, sourcePath)
		return 0, nil
	}
	httpResponseCode := ExtractHttpResponseCode(err)
	log.Debugf(ctx, "Fail to download %q from %q", destinationPath, sourcePath)
	log.Debugf(ctx, "Fail to download %q: output %s", destinationPath, out)
	log.Debugf(ctx, "Fail to download %q: httpResponseCode %d", destinationPath, httpResponseCode)
	if httpResponseCode >= 500 {
		// non-500 errors are recorded by caching service.
		// We are only interested in 500 errors coming from the caching service at the moment..
		if execMetric := metrics.GetDefaultAction(ctx); execMetric != nil {
			execMetric.Observations = append(execMetric.Observations,
				metrics.NewInt64Observation("cache_failed_response_code", int64(httpResponseCode)),
				metrics.NewStringObservation("cache_failed_source_path", sourcePath),
			)
		}
	}
	return httpResponseCode, errors.Annotate(err, "install firmware image").Err()
}
