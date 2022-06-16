// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitreferences

import (
	"context"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/testutil"
)

func TestSpan(t *testing.T) {
	Convey("With Spanner Test Database", t, func() {
		ctx := testutil.SpannerTestContext(t)

		Convey("EnsureExists", func() {
			save := func(r *GitReference) (time.Time, error) {
				commitTime, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
					return EnsureExists(ctx, r)
				})
				return commitTime.In(time.UTC), err
			}
			entry := &GitReference{
				Project: "testproject",
				GitReferenceHash: GitReferenceHash(
					"mysource.googlesource.com", "chromium/src", "refs/heads/main"),
				Hostname:   "mysource.googlesource.com",
				Repository: "chromium/src",
				Reference:  "refs/heads/main",
			}

			Convey("First EnsureExists creates entry", func() {
				commitTime, err := save(entry)
				So(err, ShouldBeNil)

				expectedEntry := &GitReference{}
				*expectedEntry = *entry
				expectedEntry.GitReferenceHash = []byte{76, 190, 164, 46, 95, 208, 176, 7}
				expectedEntry.LastIngestionTime = commitTime

				refs, err := ReadAll(span.Single(ctx))
				So(err, ShouldBeNil)
				So(refs, ShouldResemble, []*GitReference{expectedEntry})

				Convey("Repeated EnsureExists updates LastIngestionTime", func() {
					// Save again.
					commitTime, err = save(entry)
					So(err, ShouldBeNil)

					expectedEntry.LastIngestionTime = commitTime
					refs, err = ReadAll(span.Single(ctx))
					So(err, ShouldBeNil)
					So(refs, ShouldResemble, []*GitReference{expectedEntry})
				})
			})
			Convey("Hash collisions are detected", func() {
				// Hash collisions are not expected to occur in the lifetime
				// of the design, but are detected to avoid data consistency
				// issues arising. Such data consistency issues could
				// lead to a public realm seeing GitReference data for a private
				// realm (if the hash of public GitReference collies
				// with a private GitReference). In this case, we would prefer
				// to prevent write of the colliding GitReference and not ingest
				// the test results rather than overwrite the existing
				// GitReference.
				_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
					// Insert a fake colliding entry. The hash of this entry
					// does not actually match its contents, but we pretend
					// it does.
					row := map[string]interface{}{
						"Project":           "testproject",
						"GitReferenceHash":  entry.GitReferenceHash,
						"Hostname":          "othersource.googlesource.com",
						"Repository":        "otherrepo/src",
						"Reference":         "refs/heads/other",
						"LastIngestionTime": spanner.CommitTimestamp,
					}
					span.BufferWrite(ctx, spanner.InsertMap("GitReferences", row))
					return nil
				})
				So(err, ShouldBeNil)

				_, err = save(entry)
				So(err, ShouldErrLike, "GitReferenceHash collision")
			})
			Convey("Invalid entries are rejected", func() {
				Convey("Project is empty", func() {
					entry.Project = ""
					_, err := save(entry)
					So(err, ShouldErrLike, "Project does not match pattern")
				})
				Convey("Project is invalid", func() {
					entry.Project = "!invalid"
					_, err := save(entry)
					So(err, ShouldErrLike, "Project does not match pattern")
				})
				Convey("GitReferenceHash is invalid", func() {
					entry.GitReferenceHash = nil
					_, err := save(entry)
					So(err, ShouldErrLike, "GitReferenceHash is unset or inconsistent")
				})
				Convey("Hostname is empty", func() {
					entry.Hostname = ""
					_, err := save(entry)
					So(err, ShouldErrLike, "Hostname must have a length between 1 and 255")
				})
				Convey("Hostname is too long", func() {
					entry.Hostname = strings.Repeat("h", 256)
					_, err := save(entry)
					So(err, ShouldErrLike, "Hostname must have a length between 1 and 255")
				})
				Convey("Repository is empty", func() {
					entry.Repository = ""
					_, err := save(entry)
					So(err, ShouldErrLike, "Repository must have a length between 1 and 4096")
				})
				Convey("Repository is too long", func() {
					entry.Repository = strings.Repeat("r", 4097)
					_, err := save(entry)
					So(err, ShouldErrLike, "Repository must have a length between 1 and 4096")
				})
				Convey("Reference is empty", func() {
					entry.Reference = ""
					_, err := save(entry)
					So(err, ShouldErrLike, "Reference must have a length between 1 and 4096")
				})
				Convey("Reference is too long", func() {
					entry.Reference = strings.Repeat("f", 4097)
					_, err := save(entry)
					So(err, ShouldErrLike, "Reference must have a length between 1 and 4096")
				})
			})
		})
	})
}
