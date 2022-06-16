// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package gitreferences contains methods for creating and reading
// git references in Spanner.
package gitreferences

import (
	"bytes"
	"context"
	"crypto/sha256"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"
	"google.golang.org/grpc/codes"

	"infra/appengine/weetbix/internal/config"
	spanutil "infra/appengine/weetbix/internal/span"
)

// GitReference represents a row in the GitReferences table.
type GitReference struct {
	// Project is the name of the LUCI Project.
	Project string
	// GitReferenceHash is a unique key for the git reference.
	// Computed using GitReferenceHash(...).
	GitReferenceHash []byte
	// Hostname is the gittiles hostname. E.g. "chromium.googlesource.com".
	Hostname string
	// The gittiles repository name (also known as the gittiles "project").
	// E.g. "chromium/src".
	Repository string
	// The git reference name. E.g. "refs/heads/main".
	Reference string
	// Last (ingestion) time this git reference was observed.
	// This value may be out of date by up to 24 hours to allow for contention-
	// reducing strategies.
	LastIngestionTime time.Time
}

func GitReferenceHash(hostname, repository, reference string) []byte {
	result := sha256.Sum256([]byte(hostname + "\n" + repository + "\n" + reference))
	return result[:8]
}

// EnsureExists ensures the given GitReference exists in the database.
// It must be called in a Spanner transactional context.
func EnsureExists(ctx context.Context, r *GitReference) error {
	if err := validateGitReference(r); err != nil {
		return err
	}

	key := spanner.Key{r.Project, r.GitReferenceHash}
	row, err := span.ReadRow(ctx, "GitReferences", key, []string{"Hostname", "Repository", "Reference"})
	if err != nil {
		if spanner.ErrCode(err) != codes.NotFound {
			return errors.Annotate(err, "reading GitReference").Err()
		}
		// Row not found. Create it.
		row := map[string]interface{}{
			"Project":           r.Project,
			"GitReferenceHash":  r.GitReferenceHash,
			"Hostname":          r.Hostname,
			"Repository":        r.Repository,
			"Reference":         r.Reference,
			"LastIngestionTime": spanner.CommitTimestamp,
		}
		span.BufferWrite(ctx, spanner.InsertMap("GitReferences", spanutil.ToSpannerMap(row)))
	} else {
		var hostname, repository, reference string
		if err := row.Columns(&hostname, &repository, &reference); err != nil {
			return errors.Annotate(err, "reading GitReference columns").Err()
		}
		// At time of design, there are only ~500 unique GitReferences in
		// chromium in the last 90 days, so a collision in a 2^64
		// keyspace is exceedingly remote and not expected to occur
		// by chance in the life of the design. (Only if an attacker
		// gained system access and deliberately engineered a collision).
		// As a collision could allow a git reference only used in a
		// private realm to overwrite one visible from the public realm
		// (or vice-versa), to safeguard data privacy, verify no collision
		// occurred.
		if r.Hostname != hostname || r.Repository != repository || r.Reference != reference {
			return errors.Reason("GitReferenceHash collision between (%s:%s:%s) and (%s:%s:%s)",
				r.Hostname, r.Repository, r.Reference, hostname, repository, reference).Err()
		}

		// Entry exists. Perform a blind write of the LastIngestionTime.
		// This should not cause contention as we did not read
		// the LastIngestionTime cell in the same transaction.
		// See https://cloud.google.com/spanner/docs/transactions#locking.
		row := map[string]interface{}{
			"Project":           r.Project,
			"GitReferenceHash":  r.GitReferenceHash,
			"LastIngestionTime": spanner.CommitTimestamp,
		}
		span.BufferWrite(ctx, spanner.UpdateMap("GitReferences", row))
	}
	return nil
}

// validateGitReference validates that the GitReference is valid.
func validateGitReference(cr *GitReference) error {
	if !config.ProjectRe.MatchString(cr.Project) {
		return errors.Reason("Project does not match pattern %s", config.ProjectRePattern).Err()
	}
	if cr.Hostname == "" || len(cr.Hostname) > 255 {
		return errors.Reason("Hostname must have a length between 1 and 255").Err()
	}
	if cr.Repository == "" || len(cr.Repository) > 4096 {
		return errors.Reason("Repository must have a length between 1 and 4096").Err()
	}
	if cr.Reference == "" || len(cr.Reference) > 4096 {
		return errors.Reason("Reference must have a length between 1 and 4096").Err()
	}
	gitRefHash := GitReferenceHash(cr.Hostname, cr.Repository, cr.Reference)
	if !bytes.Equal(gitRefHash, cr.GitReferenceHash) {
		return errors.Reason("GitReferenceHash is unset or inconsistent, expected %v", gitRefHash).Err()
	}
	return nil
}

// ReadAll reads all git references. Provided for testing only.
func ReadAll(ctx context.Context) ([]*GitReference, error) {
	cols := []string{"Project", "GitReferenceHash", "Hostname", "Repository", "Reference", "LastIngestionTime"}
	it := span.Read(ctx, "GitReferences", spanner.AllKeys(), cols)

	var results []*GitReference
	err := it.Do(func(r *spanner.Row) error {
		var project string
		var gitRefHash []byte
		var hostname, repository, reference string
		var lastIngestionTime time.Time
		err := r.Columns(&project, &gitRefHash, &hostname, &repository, &reference, &lastIngestionTime)
		if err != nil {
			return err
		}

		ref := &GitReference{
			Project:           project,
			GitReferenceHash:  gitRefHash,
			Hostname:          hostname,
			Repository:        repository,
			Reference:         reference,
			LastIngestionTime: lastIngestionTime,
		}
		results = append(results, ref)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}
