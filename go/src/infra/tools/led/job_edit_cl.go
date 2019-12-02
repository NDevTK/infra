// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"net/url"
	"strconv"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
)

func urlTrimSplit(path string) []string {
	ret := strings.Split(strings.Trim(path, "/"), "/")
	// empty paths can return a single empty token
	if len(ret) == 1 && ret[0] == "" {
		ret = nil
	}
	return ret
}

// parseGerrit is a helper to parse ethier the url.Path or url.Fragment.
//
// toks should be [<issue>, <patchset>] or [<issue>]
func parseGerrit(p *url.URL, toks []string) (ret *bbpb.GerritChange, err error) {
	ret = &bbpb.GerritChange{}
	ret.Host = p.Host
	switch len(toks) {
	case 2:
		if ret.Patchset, err = strconv.ParseInt(toks[1], 10, 0); err != nil {
			return
		}
		fallthrough
	case 1:
		ret.Change, err = strconv.ParseInt(toks[0], 10, 0)
	default:
		err = errors.New("unrecognized URL")
	}
	return
}

func parseCrChangeListURL(clURL string) (*bbpb.GerritChange, error) {
	p, err := url.Parse(clURL)
	if err != nil {
		err = errors.Annotate(err, "URL_TO_CHANGELIST is invalid").Err()
		return nil, err
	}
	toks := urlTrimSplit(p.Path)
	if len(toks) == 0 || toks[0] == "c" || strings.Contains(p.Hostname(), "googlesource") {
		if len(toks) == 0 {
			// https://<gerrit_host>/#/c/<issue>
			// https://<gerrit_host>/#/c/<issue>/<patchset>
			toks = urlTrimSplit(p.Fragment)
			if len(toks) < 1 || toks[0] != "c" {
				return nil, errors.Reason("bad format for (old) gerrit URL: %q", clURL).Err()
			}
			toks = toks[1:] // remove "c"
		} else if len(toks) == 1 {
			// https://<gerrit_host>/<issue>
			// toks is already in the correct form
		} else {
			toks = toks[1:] // remove "c"
			// https://<gerrit_host>/c/<issue>
			// https://<gerrit_host>/c/<issue>/<patchset>
			// https://<gerrit_host>/c/<project/path>/+/<issue>
			// https://<gerrit_host>/c/<project/path>/+/<issue>/<patchset>
			for i, tok := range toks {
				if tok == "+" {
					toks = toks[i+1:]
					break
				}
			}
		}
		// toks should be [<issue>] or [<issue>, <patchset>] at this point
		ret, err := parseGerrit(p, toks)
		err = errors.Annotate(err, "bad format for gerrit URL: %q", clURL).Err()
		return ret, err
	}

	return nil, errors.Reason("Unknown changelist URL format: %q", clURL).Err()
}

// ChromiumCL edits the chromium-recipe-specific properties pertaining to
// a "tryjob" CL. These properties include things like "patch_storage", "issue",
// etc.
func (ejd *EditBBJobDefinition) ChromiumCL(patchsetURL string, atIndex int) {
	if patchsetURL == "" {
		return
	}

	ejd.tweak(func() error {
		// parse patchsetURL to see if we understand it
		cl, err := parseCrChangeListURL(patchsetURL)
		if err != nil {
			return errors.Annotate(err, "parsing changelist URL").Err()
		}

		// wipe out all the old properties
		toDel := []string{
			"blamelist", "issue", "patch_gerrit_url", "patch_issue", "patch_project",
			"patch_ref", "patch_repository_url", "patch_set", "patch_storage",
			"patchset", "repository", "rietveld", "buildbucket",
		}
		for _, key := range toDel {
			delete(ejd.bb.BbagentArgs.Build.Input.Properties.Fields, key)
		}

		gc := &ejd.bb.BbagentArgs.Build.Input.GerritChanges

		for len(*gc) < atIndex+1 {
			*gc = append(*gc, nil)
		}
		(*gc)[atIndex] = cl

		return nil
	})
}
