// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/phosphorus"
)

func TestSoftwareDefaults(t *testing.T) {
	swDeps := &test_platform.Request_Params{}
	buildDep := &test_platform.Request_Params_SoftwareDependency{
		Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{
			ChromeosBuild: "eve-release",
		},
	}
	swDeps.SoftwareDependencies = append(swDeps.SoftwareDependencies, buildDep)

	res := chromeOSBuildDependencyOrEmpty(swDeps.SoftwareDependencies)
	if res.ChromeOSBucket != "gs://chromeos-image-archive" {
		t.Errorf("defaults not returned as expected, got %v", res.ChromeOSBucket)
	}

	bucketDep := &test_platform.Request_Params_SoftwareDependency{
		Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuildGcsBucket{
			ChromeosBuildGcsBucket: "chromeos-release-test",
		},
	}

	swDeps.SoftwareDependencies = append(swDeps.SoftwareDependencies, bucketDep)
	res = chromeOSBuildDependencyOrEmpty(swDeps.SoftwareDependencies)
	if res.ChromeOSBucket != "gs://chromeos-release-test" {
		t.Errorf("defaults not returned as expected, got %v", res.ChromeOSBucket)
	}
}

func TestFirmwareBuildDependencyOrEmpty(t *testing.T) {
	cases := []struct {
		RODesired string
		RWDesired string
		Deps      []*test_platform.Request_Params_SoftwareDependency
	}{
		{
			"",
			"",
			[]*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{
						ChromeosBuild: "123",
					},
				},
			},
		},
		{
			"",
			"",
			[]*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_RoFirmwareBuild{
						RoFirmwareBuild: "",
					},
				},
			},
		},
		{
			"",
			"",
			[]*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_RwFirmwareBuild{
						RwFirmwareBuild: "",
					},
				},
			},
		},
		{
			"123",
			"",
			[]*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_RoFirmwareBuild{
						RoFirmwareBuild: "123",
					},
				},
			},
		},
		{
			"",
			"123",
			[]*test_platform.Request_Params_SoftwareDependency{
				{
					Dep: &test_platform.Request_Params_SoftwareDependency_RwFirmwareBuild{
						RwFirmwareBuild: "123",
					},
				},
			},
		},
	}
	for _, c := range cases {
		roDesired := roFirmwareBuildDependencyOrEmpty(c.Deps)
		if roDesired != c.RODesired {
			t.Errorf("ro build dependency should be %v, got %v", c.RODesired, roDesired)
		}
		rwDesired := rwFirmwareBuildDependencyOrEmpty(c.Deps)
		if rwDesired != c.RWDesired {
			t.Errorf("rw build dependency should be %v, got %v", c.RWDesired, rwDesired)
		}
	}
}

func TestWhichFirmwareLabelsToProvision(t *testing.T) {
	cases := []struct {
		RODesired string
		RWDesired string
		ROFound   string
		RWFound   string
		Expected  []string
	}{
		// none cases
		{
			RODesired: "",
			RWDesired: "",
			ROFound:   "",
			RWFound:   "",
			Expected:  []string{},
		},
		{
			RODesired: "",
			RWDesired: "",
			ROFound:   "123",
			RWFound:   "",
			Expected:  []string{},
		},
		{
			RODesired: "",
			RWDesired: "",
			ROFound:   "",
			RWFound:   "123",
			Expected:  []string{},
		},
		{
			RODesired: "",
			RWDesired: "",
			ROFound:   "123",
			RWFound:   "234",
			Expected:  []string{},
		},
		{
			RODesired: "123",
			RWDesired: "234",
			ROFound:   "123",
			RWFound:   "234",
			Expected:  []string{},
		},
		// ro only cases
		{
			RODesired: "123",
			RWDesired: "",
			ROFound:   "",
			RWFound:   "",
			Expected: []string{
				fmt.Sprintf("%s:%s", roFirmwareBuildKey, "123"),
			},
		},
		{
			RODesired: "123",
			RWDesired: "",
			ROFound:   "234",
			RWFound:   "",
			Expected: []string{
				fmt.Sprintf("%s:%s", roFirmwareBuildKey, "123"),
			},
		},
		// ro only special case
		{
			RODesired: "123",
			RWDesired: "123",
			ROFound:   "234",
			RWFound:   "234",
			Expected: []string{
				fmt.Sprintf("%s:%s", roFirmwareBuildKey, "123"),
			},
		},
		// rw only cases
		{
			RODesired: "",
			RWDesired: "123",
			ROFound:   "",
			RWFound:   "",
			Expected: []string{
				fmt.Sprintf("%s:%s", rwFirmwareBuildKey, "123"),
			},
		},
		{
			RODesired: "",
			RWDesired: "123",
			ROFound:   "",
			RWFound:   "234",
			Expected: []string{
				fmt.Sprintf("%s:%s", rwFirmwareBuildKey, "123"),
			},
		},
		// ro+rw cases
		{
			RODesired: "123",
			RWDesired: "234",
			ROFound:   "",
			RWFound:   "",
			Expected: []string{
				fmt.Sprintf("%s:%s", roFirmwareBuildKey, "123"),
				fmt.Sprintf("%s:%s", rwFirmwareBuildKey, "234"),
			},
		},
		{
			RODesired: "123",
			RWDesired: "234",
			ROFound:   "234",
			RWFound:   "123",
			Expected: []string{
				fmt.Sprintf("%s:%s", roFirmwareBuildKey, "123"),
				fmt.Sprintf("%s:%s", rwFirmwareBuildKey, "234"),
			},
		},
	}
	for _, c := range cases {
		r := whichFirmwareLabelsToProvision(c.RODesired, c.RWDesired, c.ROFound, c.RWFound)
		if !reflect.DeepEqual(r, c.Expected) {
			t.Errorf("Expected %v, but got %v", c.Expected, r)
		}
	}
}

func TestShouldRunTLSProvision(t *testing.T) {
	cases := []struct {
		Tag                          string
		DesiredProvisionableLabel    string
		ProvisionDutExperimentConfig *phosphorus.ProvisionDutExperiment
		Want                         bool
	}{
		{
			Tag:  "nil config",
			Want: false,
		},
		{
			Tag: "globally disabled",
			ProvisionDutExperimentConfig: &phosphorus.ProvisionDutExperiment{
				Enabled: false,
			},
			Want: false,
		},
		{
			Tag:                       "no allow or disallow list",
			DesiredProvisionableLabel: "octopus-release/R90-13749.0.0",
			ProvisionDutExperimentConfig: &phosphorus.ProvisionDutExperiment{
				Enabled: true,
			},
			Want: false,
		},
		{
			Tag:                       "included in allow_list",
			DesiredProvisionableLabel: "octopus-release/R90-13749.0.0",
			ProvisionDutExperimentConfig: &phosphorus.ProvisionDutExperiment{
				Enabled: true,
				CrosVersionSelector: &phosphorus.ProvisionDutExperiment_CrosVersionAllowList{
					CrosVersionAllowList: &phosphorus.ProvisionDutExperiment_CrosVersionSelector{
						Prefixes: []string{"octopus-release", "atlas-cq"},
					},
				},
			},
			Want: true,
		},
		{
			Tag:                       "not included in allow_list",
			DesiredProvisionableLabel: "octopus-release/R90-13749.0.0",
			ProvisionDutExperimentConfig: &phosphorus.ProvisionDutExperiment{
				Enabled: true,
				CrosVersionSelector: &phosphorus.ProvisionDutExperiment_CrosVersionAllowList{
					CrosVersionAllowList: &phosphorus.ProvisionDutExperiment_CrosVersionSelector{
						Prefixes: []string{"octopus-release/R87", "atlas-cq"},
					},
				},
			},
			Want: false,
		},
		{
			Tag:                       "included in disallow_list",
			DesiredProvisionableLabel: "octopus-release/R90-13749.0.0",
			ProvisionDutExperimentConfig: &phosphorus.ProvisionDutExperiment{
				Enabled: true,
				CrosVersionSelector: &phosphorus.ProvisionDutExperiment_CrosVersionDisallowList{
					CrosVersionDisallowList: &phosphorus.ProvisionDutExperiment_CrosVersionSelector{
						Prefixes: []string{"octopus-release", "atlas-cq"},
					},
				},
			},
			Want: false,
		},
		{
			Tag:                       "not included in disallow_list",
			DesiredProvisionableLabel: "octopus-release/R90-13749.0.0",
			ProvisionDutExperimentConfig: &phosphorus.ProvisionDutExperiment{
				Enabled: true,
				CrosVersionSelector: &phosphorus.ProvisionDutExperiment_CrosVersionDisallowList{
					CrosVersionDisallowList: &phosphorus.ProvisionDutExperiment_CrosVersionSelector{
						Prefixes: []string{"octopus-release/R91", "atlas-cq"},
					},
				},
			},
			Want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Tag, func(t *testing.T) {
			r := &phosphorus.PrejobRequest{
				SoftwareDependencies: []*test_platform.Request_Params_SoftwareDependency{
					{
						Dep: &test_platform.Request_Params_SoftwareDependency_ChromeosBuild{
							ChromeosBuild: c.DesiredProvisionableLabel,
						},
					},
				},
				Config: &phosphorus.Config{
					PrejobStep: &phosphorus.PrejobStep{
						ProvisionDutExperiment: c.ProvisionDutExperimentConfig,
					},
				},
			}
			if b := shouldProvisionChromeOSViaTLS(r); b != c.Want {
				t.Errorf("Incorrect response from shouldRunTLSProvision(%v): %t, want %t", r, b, c.Want)
			}
		})
	}
}
