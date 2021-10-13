package clustering

import "regexp"

// ProjectRe matches validly formed LUCI Project names.
// From https://source.chromium.org/chromium/infra/infra/+/main:luci/appengine/components/components/config/common.py?q=PROJECT_ID_PATTERN
var ProjectRe = regexp.MustCompile(`^[a-z0-9\-_]{1,40}$`)

// chunkRe matches validly formed chunk IDs.
var ChunkRe = regexp.MustCompile(`^[0-9a-f]{1,32}$`)

// algorithmRe matches validly formed clustering algorithm names.
var AlgorithmRe = regexp.MustCompile(`^[0-9a-z\-.]{1,32}$`)
