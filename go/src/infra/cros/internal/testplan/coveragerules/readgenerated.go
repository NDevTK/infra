// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package coveragerules

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/cros/internal/git"
	"io/fs"
	"os"
	"path/filepath"

	testpb "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

// ReadGenerated finds all files under dir, and parses the contents of each into
// a CoverageRuleBqRow. Every file in dir is assumed to be a JSON list of
// CoverageRule jsonprotos (note that there is no proto message representing a
// list of CoverageRules, i.e. the generated files should not actually be valid
// jsonproto, but they should be a JSON list of valid jsonproto). If any file in
// dir isn't a list of CoverageRules, an error is returned.
//
// dir must be in a git repo. This function will determine the host and project
// of the repo and the repo-relative path of each file.
func ReadGenerated(ctx context.Context, dir string) ([]*testpb.CoverageRuleBqRow, error) {
	host, project, err := git.GetRemoteHostAndProject(dir)
	if err != nil {
		return nil, err
	}

	logging.Infof(ctx, "determined dir %q has host %q and project %q", dir, host, project)

	rows := make([]*testpb.CoverageRuleBqRow, 0)
	if err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		logging.Infof(ctx, "parsing CoverageRules from %s", path)

		repoRelativePath, err := git.GetRepoRelativePath(dir, path)
		if err != nil {
			return err
		}

		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// First parse the file as a list of raw JSON objects, so we can use
		// protojson to unmarshal each object (which should be a CoverageRule)
		// later.
		unparsedRules := make([]json.RawMessage, 0)
		if err = json.Unmarshal(bytes, &unparsedRules); err != nil {
			return fmt.Errorf("failed to unmarshal JSON in %s: %w", path, err)
		}

		for _, unparsedRule := range unparsedRules {
			rule := &testpb.CoverageRule{}
			if err := protojson.Unmarshal(unparsedRule, rule); err != nil {
				return err
			}

			rows = append(rows, &testpb.CoverageRuleBqRow{
				Host:         host,
				Project:      project,
				Path:         repoRelativePath,
				CoverageRule: rule,
			})
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return rows, nil
}
