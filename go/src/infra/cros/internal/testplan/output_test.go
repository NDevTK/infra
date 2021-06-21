package testplan

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	configpb "go.chromium.org/chromiumos/config/go/api"
	"go.chromium.org/chromiumos/config/go/api/software"
	buildpb "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/payload"
	testpb "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/config/go/test/plan"
)

// buildSummary is a convenience to reduce boilerplate when creating
// SystemImage_BuildSummary in test cases.
func buildSummary(overlay, kernelVersion, chipsetOverlay, arcVersion string) *buildpb.SystemImage_BuildSummary {
	return &buildpb.SystemImage_BuildSummary{
		BuildTarget: &buildpb.SystemImage_BuildTarget{
			PortageBuildTarget: &buildpb.Portage_BuildTarget{
				OverlayName: overlay,
			},
		},
		Kernel: &buildpb.SystemImage_BuildSummary_Kernel{
			Version: kernelVersion,
		},
		Chipset: &buildpb.SystemImage_BuildSummary_Chipset{
			Overlay: chipsetOverlay,
		},
		Arc: &buildpb.SystemImage_BuildSummary_Arc{
			Version: arcVersion,
		},
	}
}

// flatConfig is a convenience to reduce boilerplate when creating FlatConfig in
// test cases.
func flatConfig(
	designConfigID,
	buildTarget string,
	fingerprintLoc configpb.HardwareFeatures_Fingerprint_Location,
) *payload.FlatConfig {
	return &payload.FlatConfig{
		HwDesignConfig: &configpb.Design_Config{
			Id: &configpb.DesignConfigId{
				Value: designConfigID,
			},
			HardwareFeatures: &configpb.HardwareFeatures{
				Fingerprint: &configpb.HardwareFeatures_Fingerprint{
					Location: fingerprintLoc,
				},
			},
		},
		SwConfig: &software.SoftwareConfig{
			SystemBuildTarget: &buildpb.SystemImage_BuildTarget{
				PortageBuildTarget: &buildpb.Portage_BuildTarget{
					OverlayName: buildTarget,
				},
			},
		},
	}
}

var buildSummaryList = &buildpb.SystemImage_BuildSummaryList{
	Values: []*buildpb.SystemImage_BuildSummary{
		buildSummary("project1", "4.14", "chipsetA", ""),
		buildSummary("project2", "4.14", "chipsetB", ""),
		buildSummary("project3", "5.4", "chipsetA", ""),
		buildSummary("project4", "3.18", "chipsetC", "R"),
		buildSummary("project5", "4.14", "chipsetA", ""),
		buildSummary("project6", "4.14", "chipsetB", "P"),
	},
}

var flatConfigList = &payload.FlatConfigList{
	Values: []*payload.FlatConfig{
		flatConfig("config1", "project1", configpb.HardwareFeatures_Fingerprint_KEYBOARD_BOTTOM_LEFT),
		flatConfig("config2", "project1", configpb.HardwareFeatures_Fingerprint_NOT_PRESENT),
		flatConfig("config3", "project3", configpb.HardwareFeatures_Fingerprint_LOCATION_UNKNOWN),
		flatConfig("config4", "project4", configpb.HardwareFeatures_Fingerprint_LEFT_SIDE),
	},
}

var dutAttributeList = &testpb.DutAttributeList{
	DutAttributes: []*testpb.DutAttribute{
		{
			Id:        &testpb.DutAttribute_Id{Value: "fingerprint_location"},
			FieldPath: "design_list.configs.hardware_features.fingerprint.location",
		},
		{
			Id:        &testpb.DutAttribute_Id{Value: "system_build_target"},
			FieldPath: "software_configs.system_build_target.portage_build_target.overlay_name",
		},
	},
}

func TestGenerateOutputs(t *testing.T) {
	t.Skipf("Temporarily disabled, see https://crbug.com/1222066")
	tests := []struct {
		name     string
		input    *plan.SourceTestPlan
		expected []*testpb.CoverageRule
	}{
		{
			name: "kernel versions",
			input: &plan.SourceTestPlan{
				Requirements: &plan.SourceTestPlan_Requirements{
					KernelVersions: &plan.SourceTestPlan_Requirements_KernelVersions{},
				},
				TestTags:        []string{"kernel"},
				TestTagExcludes: []string{"flaky"},
			},
			expected: []*testpb.CoverageRule{
				{
					Name: "kernel:3.18",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project4"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{"kernel"},
								TagExcludes: []string{"flaky"},
							},
						},
					},
				},
				{
					Name: "kernel:4.14",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project1", "project2", "project5", "project6"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{"kernel"},
								TagExcludes: []string{"flaky"},
							},
						},
					},
				},
				{
					Name: "kernel:5.4",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project3"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{"kernel"},
								TagExcludes: []string{"flaky"},
							},
						},
					},
				},
			},
		},
		{
			name: "soc families",
			input: &plan.SourceTestPlan{
				Requirements: &plan.SourceTestPlan_Requirements{
					SocFamilies: &plan.SourceTestPlan_Requirements_SocFamilies{},
				},
				TestTagExcludes: []string{"flaky"},
			},
			expected: []*testpb.CoverageRule{
				{
					Name: "soc:chipsetA",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project1", "project3", "project5"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								TagExcludes: []string{"flaky"},
							},
						},
					},
				},
				{
					Name: "soc:chipsetB",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project2", "project6"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								TagExcludes: []string{"flaky"},
							},
						},
					},
				},
				{
					Name: "soc:chipsetC",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project4"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								TagExcludes: []string{"flaky"},
							},
						},
					},
				},
			},
		},
		{
			name: "build targets and designs",
			input: &plan.SourceTestPlan{
				Requirements: &plan.SourceTestPlan_Requirements{
					KernelVersions: &plan.SourceTestPlan_Requirements_KernelVersions{},
					Fingerprint:    &plan.SourceTestPlan_Requirements_Fingerprint{},
				},
				TestTags: []string{"kernel", "fingerprint"},
			},
			expected: []*testpb.CoverageRule{
				{
					Name: "fp:present",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "fingerprint_location",
							},
							Values: []string{
								"POWER_BUTTON_TOP_LEFT",
								"KEYBOARD_BOTTOM_LEFT",
								"KEYBOARD_BOTTOM_RIGHT",
								"KEYBOARD_TOP_RIGHT",
								"RIGHT_SIDE",
								"LEFT_SIDE",
								"PRESENT",
							},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "fingerprint"},
							},
						},
					},
				},
				{
					Name: "kernel:3.18",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project4"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "fingerprint"},
							},
						},
					},
				},
				{
					Name: "kernel:4.14",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project1", "project2", "project5", "project6"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "fingerprint"},
							},
						},
					},
				},
				{
					Name: "kernel:5.4",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project3"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "fingerprint"},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple requirements",
			input: &plan.SourceTestPlan{
				Requirements: &plan.SourceTestPlan_Requirements{
					KernelVersions: &plan.SourceTestPlan_Requirements_KernelVersions{},
					SocFamilies:    &plan.SourceTestPlan_Requirements_SocFamilies{},
					ArcVersions:    &plan.SourceTestPlan_Requirements_ArcVersions{},
				},
				TestTags: []string{"kernel", "arc"},
			},
			expected: []*testpb.CoverageRule{
				{
					Name: "kernel:4.14_soc:chipsetA",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project1", "project5"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "arc"},
							},
						},
					},
				},
				{
					Name: "kernel:4.14_soc:chipsetB_arc:P",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project6"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "arc"},
							},
						},
					},
				},
				{
					Name: "kernel:5.4_soc:chipsetA",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project3"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "arc"},
							},
						},
					},
				},
				{
					Name: "kernel:3.18_soc:chipsetC_arc:R",
					DutCriteria: []*testpb.DutCriterion{
						{
							AttributeId: &testpb.DutAttribute_Id{
								Value: "system_build_target",
							},
							Values: []string{"project4"},
						},
					},
					TestSuites: []*testpb.TestSuite{
						{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"kernel", "arc"},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputs, err := generateOutputs(context.Background(), test.input, buildSummaryList, dutAttributeList)

			if err != nil {
				t.Fatalf("generateOutputs failed: %s", err)
			}

			if diff := cmp.Diff(
				test.expected,
				outputs,
				cmpopts.SortSlices(func(i, j *testpb.CoverageRule) bool {
					return i.Name < j.Name
				}),
				cmpopts.SortSlices(func(i, j string) bool {
					return i < j
				}),
			); diff != "" {
				t.Errorf("generateOutputs returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGenerateOutputsErrors(t *testing.T) {
	tests := []struct {
		name  string
		input *plan.SourceTestPlan
	}{
		{
			name: "no requirements ",
			input: &plan.SourceTestPlan{
				EnabledTestEnvironments: []plan.SourceTestPlan_TestEnvironment{
					plan.SourceTestPlan_HARDWARE,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := generateOutputs(context.Background(), test.input, buildSummaryList, dutAttributeList); err == nil {
				t.Errorf("Expected error from generateOutputs")
			}
		})
	}
}
