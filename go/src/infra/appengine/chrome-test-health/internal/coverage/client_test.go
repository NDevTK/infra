// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/appengine/chrome-test-health/api"
	"infra/appengine/chrome-test-health/internal/coverage/entities"
)

func getFakeFinditConfig() *entities.FinditConfig {
	fakeSettingsData := []byte(`
{
	"default_postsubmit_report_config": {
		"fuchsia": {
			"project": "fuchsia",
			"platform": "fuchsia",
			"host": "fuchsia.googlesource.com",
			"ref": "refs/heads/main"
		},
		"chromium": {
			"project": "chromium/src",
			"platform": "linux",
			"host": "chromium.googlesource.com",
			"ref": "refs/heads/main"
		},
		"cros": {
			"project": "chromiumos/platform2",
			"platform": "cros",
			"host": "chromium.googlesource.com",
			"ref": "refs/heads/main"
		}
	}
}`)

	return &entities.FinditConfig{
		CodeCoverageSettings: fakeSettingsData,
	}
}

func getFakeFinditConfigWithoutAnyProject() *entities.FinditConfig {
	fakeSettingsData := []byte(`
{
	"default_postsubmit_report_config": {}
}`)

	return &entities.FinditConfig{
		CodeCoverageSettings: fakeSettingsData,
	}
}

func TestGetProjectConfig(t *testing.T) {
	t.Parallel()
	Convey(`Should have valid "FinditConfig" entity`, t, func() {
		client := Client{}
		ctx := context.Background()
		Convey(`Invalid "CodeCoverageSettings" JSON`, func() {
			fakeFinditConfig := entities.FinditConfig{
				CodeCoverageSettings: []byte(""),
			}
			config, err := client.getProjectConfig(ctx, &fakeFinditConfig, "chromium")
			So(config, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Missing "default_postsubmit_report_config" property`, func() {
			fakeFinditConfig := entities.FinditConfig{
				CodeCoverageSettings: []byte("{}"),
			}
			config, err := client.getProjectConfig(ctx, &fakeFinditConfig, "chromium")
			So(config, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Missing project from "default_postsubmit_report_config" property`, func() {
			fakeFinditConfig := getFakeFinditConfigWithoutAnyProject()
			config, err := client.getProjectConfig(ctx, fakeFinditConfig, "chromium")
			So(config, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrInternalServerError)
		})
		Convey(`Valid "FinditConfig" entity`, func() {
			fakeFinditConfig := getFakeFinditConfig()
			var config *api.GetProjectDefaultConfigResponse
			config, err := client.getProjectConfig(ctx, fakeFinditConfig, "chromium")
			So(config, ShouldResemble, &api.GetProjectDefaultConfigResponse{
				Host:     "chromium.googlesource.com",
				Platform: "linux",
				Project:  "chromium/src",
				Ref:      "refs/heads/main",
			})
			So(err, ShouldBeNil)
		})
	})
}
