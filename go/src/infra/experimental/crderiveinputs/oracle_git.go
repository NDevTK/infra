// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/exec"
)

type gitRepo struct {
	ctx context.Context

	repo     string
	repopath string
	cachekey string

	fetchMu  sync.RWMutex
	fetchMap map[string]error
}

func (g *gitRepo) run(args ...string) error {
	c := exec.CommandContext(g.ctx, "git", args...)
	c.Dir = g.repopath
	if err := c.Run(); err != nil {
		return errors.Annotate(err, "gitRepo{%q}.run(%q)", g.repo, args).Err()
	}
	return nil
}

func (g *gitRepo) output(args ...string) (string, error) {
	c := exec.CommandContext(g.ctx, "git", args...)
	c.Dir = g.repopath
	out, err := c.Output()
	if err != nil {
		return "", errors.Annotate(err, "gitRepo{%q}.gitOutput(%q)", g.repo, args).Err()
	}
	return string(out), err
}

func (g *gitRepo) ensureFetched(commit string) error {
	// we either need to fetch it, or this blob just doesn't exist
	g.fetchMu.RLock()
	err, ok := g.fetchMap[commit]
	g.fetchMu.RUnlock()
	if ok {
		return err
	}

	g.fetchMu.Lock()
	defer g.fetchMu.Unlock()
	if err := g.fetchMap[commit]; err != nil {
		return err
	}

	err = g.run("rev-list", "-1", "--missing=allow-any", fmt.Sprintf("%s^{commit}", commit), "--")
	if err != nil {
		// we need to fetch it - if we had fetched it before, the rev-list returns
		// without error.
		Logger.Infof("SLOW!!! - Fetching %s %s", g.repo, commit)
		err = g.run("fetch", g.repo, "--depth=1", "--filter=blob:none", commit)
	}

	g.fetchMap[commit] = err
	return err
}

func (g *gitRepo) catblob(commit, path string) (string, error) {
	if err := g.ensureFetched(commit); err != nil {
		return "", err
	}
	ret, err := g.output("cat-file", "blob", fmt.Sprintf("%s:%s", commit, path))
	return ret, err
}

func (o *Oracle) gitRepo(repo string) *gitRepo {
	// Homogenize googlesource repo URLs.
	if strings.Contains(repo, ".googlesource.com") && strings.HasSuffix(repo, ".git") {
		repo = repo[:len(repo)-len(".git")]
	}

	key := sha1.Sum([]byte(repo))
	cachekey := path.Base(repo) + "-" + hex.EncodeToString(key[:])

	o.repoMapMu.Lock()
	defer o.repoMapMu.Unlock()

	ret := o.repoMap[cachekey]
	if ret == nil {
		// We prefix the hash with the repo name to make tab completion and `ls` a bit
		// nicer when poking around the cache state.
		path := filepath.Join(o.gitpath, cachekey)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0700); err != nil {
				panic(err)
			}
			c := exec.CommandContext(o.ctx, "git", "-C", path, "init", "--bare")
			if err := c.Run(); err != nil {
				panic(errors.Annotate(err, "Oracle.gitDir(%q)", repo).Err())
			}
		}
		ret = &gitRepo{
			ctx:      o.ctx,
			repo:     repo,
			repopath: path,
			cachekey: cachekey,
			fetchMap: map[string]error{},
		}
		o.repoMap[cachekey] = ret
	}
	return ret
}
