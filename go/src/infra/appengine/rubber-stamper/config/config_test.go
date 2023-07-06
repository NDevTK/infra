// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/gae/impl/memory"
)

func TestConfig(t *testing.T) {
	Convey("loads config and updates context", t, func() {
		sampleCfg := &Config{
			HostConfigs: map[string]*HostConfig{
				"test-host": {
					RepoConfigs: map[string]*RepoConfig{
						"dummy": {
							BenignFilePattern: &BenignFilePattern{
								Paths: []string{"whitespace.txt", "a/*.txt"},
							},
						},
					},
				},
			},
		}

		c := memory.Use(context.Background())
		SetTestConfig(c, sampleCfg)

		cfg, err := Get(c)
		So(err, ShouldBeNil)

		_, ok := cfg.HostConfigs["test-host"]
		So(ok, ShouldEqual, true)
	})
}

func TestIsRepoRegexpConfigMatch(t *testing.T) {
	Convey("Works", t, func() {
		sampleCfg := &RepoConfig{
			BenignFilePattern: &BenignFilePattern{
				Paths: []string{"whitespace.txt", "a/*.txt"},
			},
		}
		Convey("Should return the matched config when there is one", func() {
			rrcfgs := []*HostConfig_RepoRegexpConfigPair{
				{
					Key:   "^dummy/dummy-.*$",
					Value: sampleCfg,
				},
				{
					Key:   "^invalid/.*$",
					Value: nil,
				},
			}
			So(RetrieveRepoRegexpConfig(context.Background(), "dummy/dummy-valid", rrcfgs), ShouldEqual, sampleCfg)
		})
		Convey("Should return nil when no matched config", func() {
			rrcfgs := []*HostConfig_RepoRegexpConfigPair{
				{
					Key:   "^dummy/dummy-.*$",
					Value: sampleCfg,
				},
			}
			So(RetrieveRepoRegexpConfig(context.Background(), "dummy-valid", rrcfgs), ShouldEqual, nil)
		})
	})
}
