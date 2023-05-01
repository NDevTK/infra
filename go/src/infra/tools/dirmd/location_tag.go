// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmd

import (
	dirmdpb "infra/tools/dirmd/proto"

	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/resultdb/pbutil"
	resultpb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
)

func createTag(md *dirmdpb.Metadata) []*resultpb.StringPair {
	var tags []*resultpb.StringPair

	if md.GetMonorail().GetComponent() != "" {
		tags = append(tags, pbutil.StringPair("monorail_component", md.Monorail.Component))
	}

	if md.GetOs() != dirmdpb.OS_OS_UNSPECIFIED {
		tags = append(tags, pbutil.StringPair("os", md.Os.String()))
	}

	if md.GetTeamEmail() != "" {
		tags = append(tags, pbutil.StringPair("team_email", md.TeamEmail))
	}

	if len(md.GetResultdb().GetTags()) > 0 {
		tags = append(tags, pbutil.FromStrpairMap(strpair.ParseMap(md.Resultdb.Tags))...)
	}

	return tags
}

// ToLocationTags converts all dir metadata to test location tags.
func ToLocationTags(mapping *Mapping) (*sinkpb.LocationTags_Repo, error) {
	dirs := map[string]*sinkpb.LocationTags_Dir{}
	for k, md := range mapping.Dirs {
		tags := createTag(md)

		dirs[k] = &sinkpb.LocationTags_Dir{
			Tags: tags,
		}
	}

	files := map[string]*sinkpb.LocationTags_File{}
	for k, md := range mapping.Files {
		tags := createTag(md)

		files[k] = &sinkpb.LocationTags_File{
			Tags: tags,
		}
	}
	return &sinkpb.LocationTags_Repo{
		Dirs:  dirs,
		Files: files,
	}, nil
}
