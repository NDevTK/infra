// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ethernethook

import (
	"context"
	"regexp"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/iterator"
)

// NewRegexDownloader creates an object that manages the common pattern of downloading files matching certain regexes
// and a query. For example, grabbing both `dmesg.gz` and `status.log` out.
func NewRegexDownloader(bucket string, query *storage.Query, inputPatterns [][]string) (*regexDownloader, error) {
	var out regexDownloader
	out.bucket = bucket
	out.query = query
	var mErr errors.MultiError
	for i, kv := range inputPatterns {
		if ok := len(kv) == 2; !ok {
			return nil, errors.Reason("key-value pair #%d has length#%d", i, len(kv)).Err()
		}
		pattern := kv[1]
		r, cErr := regexp.Compile(pattern)
		if cErr != nil {
			mErr = append(mErr, cErr)
			continue
		}
		out.patterns = append(out.patterns, r)
	}
	if len(mErr) > 0 {
		return nil, mErr
	}
	return &out, nil
}

// regexDownloader manages a download attempt.
//
// The patterns themselves must already be compiled, this achieves two things:
// 1) If the patterns compile, we know that they must be valid regular expressions, thus simplifying the interface.
// 2) We avoid the probably-negligble overhead of constantly recompiling the regex.
//
// Note that it contains the storage attributes, which are the result of scanning, directly on itself.
// It also has a scan limit which limits the total number of items scanned; it is NOT (directly) a limit
// on the number of Attrs.
type regexDownloader struct {
	bucket   string
	query    *storage.Query
	patterns []*regexp.Regexp
	// Attrs are the stored results of a scan query.
	Attrs     []*storage.ObjectAttrs
	scanLimit int
}

// ScanLimit gets the true scan limit, applying the completely reasonable and not at all arbitrary default of
// 10000 if no explicit scan limit is provided.
func (d regexDownloader) ScanLimit() int {
	if d.scanLimit <= 0 {
		return 10000
	}
	return d.scanLimit
}

// FindPaths finds the paths and attaches them to the downloader.
func (d *regexDownloader) FindPaths(ctx context.Context, e *extendedGSClient) error {
	var state LsState
	it := e.Ls(ctx, d.bucket, d.query)
	tally := 0
	for it(&state) {
		if tally > d.ScanLimit() {
			return errors.Reason("scan limit %d exceeded", d.ScanLimit()).Err()
		}
		tally++
		name := e.ExpandName(d.bucket, state.Attrs)
		for _, pattern := range d.patterns {
			if pattern.MatchString(name) {
				d.Attrs = append(d.Attrs, state.Attrs)
			}
		}
	}
	if state.Err == nil || errors.Is(state.Err, iterator.Done) {
		return nil
	}
	return errors.Annotate(state.Err, "regex downloader find paths").Err()
}

// Len returns the length of the stored attributes. A length greater than zero indicates that we succesffully scanned the area described by the query.
func (d *regexDownloader) Len() int {
	return len(d.Attrs)
}
