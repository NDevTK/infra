// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bootstrap

import "go.chromium.org/luci/common/errors"

type dependencyPropertiesFileNotFoundTag struct {
	Key errors.TagKey
}

type dependencyPropertiesFileNotFoundDetails struct {
	propsFile    string
	commit       *GitilesCommit
	upstreamRepo string
}

func (d dependencyPropertiesFileNotFoundTag) With(propsFile string, commit *GitilesCommit, upstreamRepo string) errors.TagValue {
	return errors.TagValue{
		Key: d.Key,
		Value: dependencyPropertiesFileNotFoundDetails{
			propsFile:    propsFile,
			commit:       commit,
			upstreamRepo: upstreamRepo,
		},
	}
}

func (d dependencyPropertiesFileNotFoundTag) In(err error) (propsFile string, commit *GitilesCommit, upstreamRepo string, ok bool) {
	v, ok := errors.TagValueIn(d.Key, err)
	if ok {
		details := v.(dependencyPropertiesFileNotFoundDetails)
		propsFile = details.propsFile
		commit = details.commit
		upstreamRepo = details.upstreamRepo
	}
	return
}

var (
	// DependencyPropertiesFileNotFound indicates that the properties file does not exist in the
	// dependency project at the revision pinned by the top level project.
	DependencyPropertiesFileNotFound = dependencyPropertiesFileNotFoundTag{Key: errors.NewTagKey("the properties file does not exist in the dependency project")}
	// PatchRejected indicates that some portion of a patch was rejected.
	PatchRejected = errors.BoolTag{Key: errors.NewTagKey("the patch could not be applied")}
)
