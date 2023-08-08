// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package gs exports helpers to upload log data to Google Storage.
package gs

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/common/sync/parallel"
	"google.golang.org/api/googleapi"
)

// DirWriter exposes methods to write a local directory to Google Storage.
type DirWriter struct {
	client               gsClient
	maxConcurrentUploads int
	retryIterator        retry.Iterator
}

// gsClient is a Google Storage client.
//
// This interface is a subset of the gs.Client interface.
type gsClient interface {
	NewWriter(p gs.Path) (gs.Writer, error)
}

// transientErrorRetryIterator chooses a delay for all transient errors.
// It stops on success or a non-transient error.
type transientErrorRetryIterator struct {
	impl retry.Iterator
}

// newTransientErrorRetryIterator creates a new transient error retry iterator.
func newTransientErrorRetryIterator(retryLimit int) *transientErrorRetryIterator {
	return &transientErrorRetryIterator{
		impl: &concurrencySafeRetryIterator{
			i: &retry.ExponentialBackoff{
				Limited: retry.Limited{
					Delay:   100 * time.Millisecond,
					Retries: retryLimit,
				},
				MaxDelay:   30 * time.Second,
				Multiplier: 2,
			},
		},
	}
}

// Next picks a duration to wait after an error. Returns stop if and only if the error is nil or transient.
func (i *transientErrorRetryIterator) Next(ctx context.Context, e error) time.Duration {
	// The first few checks are thread-safe and can happen outside the core of the retry iterator
	// that's governed by the mutex.
	if e == nil {
		return retry.Stop
	}
	if cloudErr, code := extractCloudErrorCode(e); cloudErr != nil && code == 403 {
		return retry.Stop
	}
	select {
	case <-ctx.Done():
		return retry.Stop
	default:
		return i.impl.Next(ctx, e)
	}
}

// NewDirWriter creates an object which can write a directory and its subdirectories to the given Google Storage path.
func NewDirWriter(client gsClient, maxConcurrentUploads, retryLimit int) *DirWriter {
	return &DirWriter{
		client:               client,
		maxConcurrentUploads: maxConcurrentUploads,
		retryIterator:        newTransientErrorRetryIterator(retryLimit),
	}
}

func verifyPaths(localPath string, gsPath string) error {
	problems := []string{}
	if _, err := os.Stat(localPath); err != nil {
		problems = append(problems, fmt.Sprintf("invalid local path (%s)", localPath))
	}
	if _, err := url.Parse(gsPath); err != nil {
		problems = append(problems, fmt.Sprintf("invalid GS path (%s)", gsPath))
	}
	if len(problems) > 0 {
		return errors.Reason("path errors: %s", strings.Join(problems, ", ")).Err()
	}
	return nil
}

// WriteDir writes a local directory to Google Storage.
//
// If ctx is canceled, WriteDir() returns after completing in-flight uploads,
// skipping remaining contents of the directory and returns ctx.Err().
func (w *DirWriter) WriteDir(ctx context.Context, srcDir string, dstDir gs.Path) error {
	logging.Debugf(ctx, "Writing %s and subtree to %s.", srcDir, dstDir)
	if err := verifyPaths(srcDir, string(dstDir)); err != nil {
		return err
	}

	files, merr := discoverFiles(srcDir, dstDir)

	var terr error
	err := parallel.WorkPool(w.maxConcurrentUploads, func(items chan<- func() error) {
		for _, f := range files {
			// Create a loop-local variable for capture in the lambda.
			f := f
			item := func() error {
				// Check the context timeout when trying to upload.
				select {
				case <-ctx.Done():
					terr = ctx.Err()
					// Context error will be added separate.
					return nil
				default:
					return w.writeOne(ctx, f)
				}
			}
			// Check the context timeout when adding files to the stack.
			select {
			case <-ctx.Done():
				terr = ctx.Err()
				return
			default:
				items <- item
			}
		}
	})
	if err != nil {
		merr = append(merr, err)
	}
	if terr != nil {
		merr = append(merr, err)
	}
	if len(merr) > 0 {
		return errors.Annotate(merr, "writing dir %s to %s", srcDir, dstDir).Err()
	}
	return nil
}

func discoverFiles(srcDir string, dstDir gs.Path) ([]*file, errors.MultiError) {
	var merr errors.MultiError
	files := []*file{}
	if err := filepath.Walk(srcDir, func(src string, info os.FileInfo, err error) error {
		// Continue walking the directory tree on errors so that we upload as
		// many files as possible.
		if err != nil {
			merr = append(merr, errors.Annotate(err, "list files to upload: %s", src).Err())
			return nil
		}
		relPath, err := filepath.Rel(srcDir, src)
		if err != nil {
			merr = append(merr, errors.Annotate(err, "writing from %s to %s", src, dstDir).Err())
			return nil
		}
		files = append(files, &file{
			Src:  src,
			Dest: dstDir.Concat(relPath),
			Info: info,
		})
		return nil
	}); err != nil {
		panic(fmt.Sprintf("Directory walk leaked error: %s", err))
	}
	return files, merr
}

func (w *DirWriter) writeOne(ctx context.Context, f *file) error {
	err := f.Write(ctx, w.client)
	for err != nil {
		d := w.retryIterator.Next(ctx, err)
		if d == retry.Stop {
			break
		}
		logging.Warningf(ctx, "%s failed upload: %s. Will retry after %s", f.Src, err, d.String())
		// This sleep implies that the worker goroutine trying to upload this
		// file will block. Because we use parallel.WorkPool(), this means that
		// one of the fix number of concurrent goroutines will be blocked.
		//
		// This is intentional: Most errors are due to transient service
		// degradation in Google Storage. Blocking the worker goroutines ensures
		// that our overall upload throughput is throttled in case of such
		// transient errors.
		time.Sleep(d)
		err = f.Write(ctx, w.client)
	}
	return err
}

type file struct {
	Src  string
	Dest gs.Path
	Info os.FileInfo
}

func (f *file) Write(ctx context.Context, client gsClient) error {
	if f.Info.IsDir() {
		return nil
	}
	if skip, reason := shouldSkipUpload(f.Info); skip {
		logging.Debugf(ctx, "Skipped %s because: %s.", f.Src, reason)
		return nil
	}

	r, err := os.Open(f.Src)
	if err != nil {
		return errors.Annotate(err, "writing from %s to %s", f.Src, f.Dest).Err()
	}
	defer r.Close()

	writer, err := client.NewWriter(f.Dest)
	if err != nil {
		return errors.Annotate(err, "writing from %s to %s", f.Src, f.Dest).Err()
	}
	// Ignore errors as we may have already closed writer by the time this runs.
	defer writer.Close()

	bs := make([]byte, f.Info.Size())
	if _, err = r.Read(bs); err != nil {
		return errors.Annotate(err, "writing from %s to %s", f.Src, f.Dest).Err()
	}
	n, err := writer.Write(bs)
	if err != nil {
		return errors.Annotate(err, "writing from %s to %s", f.Src, f.Dest).Err()
	}
	if int64(n) != f.Info.Size() {
		return errors.Reason("length written to %s does not match source file size", f.Dest).Err()
	}
	err = writer.Close()
	if err != nil {
		return errors.Annotate(err, "writer for %s failed to close", f.Dest).Err()
	}
	return nil
}

// shouldSkipUpload determines if a particular file should be skipped.
//
// Also returns a reason for skipping the file.
func shouldSkipUpload(i os.FileInfo) (bool, string) {
	if i.Mode()&os.ModeType == 0 {
		return false, ""
	}

	switch {
	case i.Mode()&os.ModeSymlink == os.ModeSymlink:
		return true, "file is a symlink"
	case i.Mode()&os.ModeDevice == os.ModeDevice:
		return true, "file is a device"
	case i.Mode()&os.ModeNamedPipe == os.ModeNamedPipe:
		return true, "file is a named pipe"
	case i.Mode()&os.ModeSocket == os.ModeSocket:
		return true, "file is a unix domain socket"
	case i.Mode()&os.ModeIrregular == os.ModeIrregular:
		return true, "file is an irregular file of unknown type"
	default:
		return true, "file is a non-file of unknown type"
	}
}

type concurrencySafeRetryIterator struct {
	i retry.Iterator
	m sync.Mutex
}

func (r *concurrencySafeRetryIterator) Next(ctx context.Context, err error) time.Duration {
	r.m.Lock()
	defer r.m.Unlock()
	return r.i.Next(ctx, err)
}

// errorExtractionLimit is the maximum number of times that we can unwrap an error.
// We set this to a finite value out of paranoia.
const errorExtractionLimit = 1000

// extractCloudErrorCode takes an error that wraps a *googleapi.Error (possibly an
// arbitrary number of times less than the errorExtractionLimit) and returns the *googleapi.Error and the exit code contained therein.
//
// extractCloudErrorCode returns (nil, 0) if and only if e is not and does not contain a *googleapi.Error.
func extractCloudErrorCode(e error) (*googleapi.Error, int) {
	cur := e
	for i := 1; i <= errorExtractionLimit; i++ {
		switch v := cur.(type) {
		case *googleapi.Error:
			return v, v.Code
		}
		cur = errors.Unwrap(cur)
	}
	return nil, 0
}
