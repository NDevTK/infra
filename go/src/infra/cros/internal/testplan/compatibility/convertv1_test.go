// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package compatibility_test

import (
	"context"
	"math/rand"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	testpb "go.chromium.org/chromiumos/config/go/test/api"
	test_api_v1 "go.chromium.org/chromiumos/config/go/test/api/v1"
	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/lab"
	"go.chromium.org/chromiumos/infra/proto/go/testplans"
	bbpb "go.chromium.org/luci/buildbucket/proto"

	"infra/cros/internal/testplan/compatibility"
)

// newStruct is a convenience method to build a structpb.Struct from a map of
// string -> interface. For example:
//
// newStruct(t, map[string]interface{}{
// "a": 1, "b": []interface{}{"c", "d"}
// })
//
// Any errors will be passed to t.Fatal. See structpb.NewValue for more info
// on how Go interfaces are converted to structpb.Struct.
func newStruct(t *testing.T, fields map[string]interface{}) *structpb.Struct {
	s := &structpb.Struct{Fields: map[string]*structpb.Value{}}

	for key, val := range fields {
		valPb, err := structpb.NewValue(val)
		if err != nil {
			t.Fatal(err)
		}

		s.Fields[key] = valPb
	}

	return s
}

var hwTestPlans = []*test_api_v1.HWTestPlan{
	{
		CoverageRules: []*testpb.CoverageRule{
			// Criticality not set will default to true.
			{
				TestSuites: []*testpb.TestSuite{
					{
						Spec: &testpb.TestSuite_TestCaseIds{
							TestCaseIds: &testpb.TestCaseIdList{
								TestCaseIds: []*testpb.TestCase_Id{
									{
										Value: "suite1",
									},
									{
										Value: "suite2",
									},
									// Add a suite twice, it should be de-duped.
									{
										Value: "suite1",
									},
								},
							},
						},
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								// "boardA" will be chosen, since it is critical and has the lowest priority.
								Values: []string{"boardC", "boardA", "boardB", "non-critical-board"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_QUOTA"},
							},
						},
					},
				},
			},
			// Specified the number of shards.
			{
				TestSuites: []*testpb.TestSuite{
					{
						Name: "suite3",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{`"group:somegroup"`},
								TagExcludes: []string{"informational"},
							},
						},
						TotalShards: 5,
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"boardA"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-model",
								},
								Values: []string{"model1"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "misc-license",
								},
								Values: []string{"LICENSE_TYPE_WINDOWS_10_PRO"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "misc-license",
								},
								Values: []string{"LICENSE_TYPE_MS_OFFICE_STANDARD"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_QUOTA"},
							},
						},
					},
				},
				Critical: &wrapperspb.BoolValue{Value: false},
			},
			// Critical rule running on a non-critical builder.
			{
				TestSuites: []*testpb.TestSuite{
					{
						Spec: &testpb.TestSuite_TestCaseIds{
							TestCaseIds: &testpb.TestCaseIdList{
								TestCaseIds: []*testpb.TestCase_Id{
									{
										Value: "suite-with-board-variant",
									},
								},
							},
						},
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"boardA"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-model",
								},
								Values: []string{"model1"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_QUOTA"},
							},
						},
						ProvisionConfig: &testpb.ProvisionConfig{
							BoardVariant: "kernelnext",
						},
					},
				},
				Critical:  &wrapperspb.BoolValue{Value: true},
				RunViaCft: true,
			},
			{
				TestSuites: []*testpb.TestSuite{
					{
						Spec: &testpb.TestSuite_TestCaseIds{
							TestCaseIds: &testpb.TestCaseIdList{
								TestCaseIds: []*testpb.TestCase_Id{
									{
										Value: "asan-suite",
									},
								},
							},
						},
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"boardA"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_QUOTA"},
							},
						},
						ProvisionConfig: &testpb.ProvisionConfig{
							Profile: "asan",
						},
					},
				},
				Critical:               &wrapperspb.BoolValue{Value: true},
				RunViaCft:              true,
				EnableAutotestSharding: true,
			},
			{
				TestSuites: []*testpb.TestSuite{
					{
						Spec: &testpb.TestSuite_TestCaseIds{
							TestCaseIds: &testpb.TestCaseIdList{
								TestCaseIds: []*testpb.TestCase_Id{
									{
										Value: "suite1",
									},
								},
							},
						},
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								// boardB doesn't contain any testable artifacts
								// so this should be skipped.
								Values: []string{"boardB"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_QUOTA"},
							},
						},
					},
				},
				Critical:               &wrapperspb.BoolValue{Value: true},
				EnableAutotestSharding: true,
			},
			{
				TestSuites: []*testpb.TestSuite{
					{
						Spec: &testpb.TestSuite_TestCaseIds{
							TestCaseIds: &testpb.TestCaseIdList{
								TestCaseIds: []*testpb.TestCase_Id{
									{
										Value: "suite1",
									},
								},
							},
						},
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								// board-bad-containers failed to build CFT
								// containers, so should not run this suite
								// (RunViaCft is set true).
								Values: []string{"board-bad-containers"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_QUOTA"},
							},
						},
					},
				},
				Critical:  &wrapperspb.BoolValue{Value: true},
				RunViaCft: true,
			},
			// Multi-dut tests.
			{
				TestSuites: []*testpb.TestSuite{
					{
						Name: "suite4-multi-dut",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{`"group:somegroup"`},
								TagExcludes: []string{"informational"},
							},
						},
						TotalShards: 1,
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"boardA", "boardB"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"DUT_POOL_MULTI_DUT"},
							},
						},
					},
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"boardCompanionA", "boardCompanionB"},
							},
						},
					},
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"pixelA", "pixelB"},
							},
						},
						ProvisionConfig: &testpb.ProvisionConfig{
							Companion: &testpb.CompanionConfig{
								Config: &testpb.CompanionConfig_Android_{
									Android: &testpb.CompanionConfig_Android{
										GmsCorePackage: "latest-stable",
									},
								},
							},
						},
					},
				},
				Critical: &wrapperspb.BoolValue{Value: false},
			},
		},
	},
}

var vmTestPlans = []*test_api_v1.VMTestPlan{
	{
		CoverageRules: []*testpb.CoverageRule{
			{
				Name: "vmrule",
				TestSuites: []*testpb.TestSuite{
					{
						Name: "tast_vm_suite1",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{"\"group:mainline\"", "\"dep:depA\""},
								TagExcludes: []string{"informational"},
							},
						},
						TotalShards: 1,
					},
					{
						Name: "tast_gce_suite2",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"\"group:mainline\"", "informational"},
							},
						},
						TotalShards: 2,
					},
					{
						Name: "tast_vm_hwsec_cq",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:             []string{"group:mainline", "dep:depB"},
								TagExcludes:      []string{"informational"},
								TestNames:        []string{"hwsec.*", "tast.cryptohome.*"},
								TestNameExcludes: []string{"firmware.*", "tast.arc.*"},
							},
						},
						TotalShards: 1,
					},
					// Add suites twice, they should be de-duped.
					{
						Name: "tast_vm_suite1",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{"\"group:mainline\"", "\"dep:depA\""},
								TagExcludes: []string{"informational"},
							},
						},
						TotalShards: 1,
					},
					{
						Name: "tast_gce_suite2",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags: []string{"\"group:mainline\"", "informational"},
							},
						},
						TotalShards: 2,
					},
					{
						Name: "tast_vm_hwsec_cq",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:             []string{"group:mainline", "dep:depB"},
								TagExcludes:      []string{"informational"},
								TestNames:        []string{"hwsec.*", "cryptohome.*"},
								TestNameExcludes: []string{"firmware.*"},
							},
						},
						TotalShards: 1,
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"vmboardA", "vmboardB"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"VM_POOL"},
							},
						},
					},
				},
				Critical: &wrapperspb.BoolValue{Value: true},
			},
			{
				Name: "vmrule-with-variant",
				TestSuites: []*testpb.TestSuite{
					{
						Name: "tast_vm_suite1",
						Spec: &testpb.TestSuite_TestCaseTagCriteria_{
							TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{"\"group:mainline\"", "\"dep:depA\""},
								TagExcludes: []string{"informational"},
							},
						},
						TotalShards: 3,
					},
				},
				DutTargets: []*testpb.DutTarget{
					{
						Criteria: []*testpb.DutCriterion{
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "attr-program",
								},
								Values: []string{"vmboardA"},
							},
							{
								AttributeId: &testpb.DutAttribute_Id{
									Value: "swarming-pool",
								},
								Values: []string{"VM_POOL"},
							},
						},
						ProvisionConfig: &testpb.ProvisionConfig{
							BoardVariant: "arc-r",
						},
					},
				},
			},
		},
	},
}

func serializeOrFatal(t *testing.T, m proto.Message) *testplans.ProtoBytes {
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	return &testplans.ProtoBytes{SerializedProto: b}
}

func getSerializedBuilds(t *testing.T) []*testplans.ProtoBytes {
	build1 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builderA",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "boardA",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathA",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_YES,
	}

	build2 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builderB",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "boardB",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathB",
					"files_by_artifact": map[string]interface{}{
						"testartifact": []interface{}{"file1", "file2"},
						// A test artifact is an empty list, this should be ignored.
						"IMAGE_ZIP": []interface{}{},
						// A test artifact is not a list, this should also be ignored.
						"TEST_UPDATE_PAYLOAD": 123,
					},
				},
			}),
		},
		Critical: bbpb.Trinary_YES,
	}

	build3 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builderC",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "boardC",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathC",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
	}

	build4 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "non-critical-builder",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "non-critical-board",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspath",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_NO,
	}

	build5 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "pointless-build",
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"pointless_build": true,
			}),
		},
	}

	build6 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "no-build-target-build",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"other_input_prop": 12,
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathB",
					"files_by_artifact": map[string]interface{}{
						"testartifact": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
	}

	variantBuild := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builderA-kernelnext",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "boardA-kernelnext",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathA-kernelnext",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_NO,
	}

	vmBuild := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-vmBuilderA",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "vmboardA",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathA",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_YES,
	}

	vmBuildWithVariant := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-vmBuilderA-arc-r",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "vmboardA-arc-r",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathA-arc-r",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_YES,
	}

	asanBuild := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builderA-asan",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "boardA",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathA-asan",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_YES,
	}

	vmOptimizedBuild := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builderA-vm-optimized",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "boardA",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspathA",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
			}),
		},
		Critical: bbpb.Trinary_YES,
	}

	buildWithFailedContainers := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Builder: "cq-builder-bad-containers",
		},
		Input: &bbpb.Build_Input{
			Properties: newStruct(t, map[string]interface{}{
				"build_target": map[string]interface{}{
					"name": "board-bad-containers",
				},
			}),
		},
		Output: &bbpb.Build_Output{
			Properties: newStruct(t, map[string]interface{}{
				"artifacts": map[string]interface{}{
					"gs_bucket": "testgsbucket",
					"gs_path":   "testgspath",
					"files_by_artifact": map[string]interface{}{
						"IMAGE_ZIP": []interface{}{"file1", "file2"},
					},
				},
				"container_building_failed": true,
			}),
		},
	}

	return []*testplans.ProtoBytes{
		serializeOrFatal(t, build1),
		serializeOrFatal(t, build2),
		serializeOrFatal(t, build3),
		serializeOrFatal(t, build4),
		serializeOrFatal(t, build5),
		serializeOrFatal(t, build6),
		serializeOrFatal(t, variantBuild),
		serializeOrFatal(t, vmBuild),
		serializeOrFatal(t, vmBuildWithVariant),
		serializeOrFatal(t, asanBuild),
		serializeOrFatal(t, vmOptimizedBuild),
		serializeOrFatal(t, buildWithFailedContainers),
	}
}

var builderConfigs = &chromiumos.BuilderConfigs{
	BuilderConfigs: []*chromiumos.BuilderConfig{
		{
			Id: &chromiumos.BuilderConfig_Id{
				Name: "cq-builderA-asan",
			},
			Build: &chromiumos.BuilderConfig_Build{
				PortageProfile: &chromiumos.BuilderConfig_Build_PortageProfile{
					Profile: "asan",
				},
			},
		},
		{
			Id: &chromiumos.BuilderConfig_Id{
				Name: "cq-builderA-vm-optimized",
			},
			Build: &chromiumos.BuilderConfig_Build{
				PortageProfile: &chromiumos.BuilderConfig_Build_PortageProfile{
					Profile: "vm-optimized",
				},
			},
		},
	},
}

var dutAttributeList = &testpb.DutAttributeList{
	DutAttributes: []*testpb.DutAttribute{
		{
			Id: &testpb.DutAttribute_Id{
				Value: "attr-program",
			},
			Aliases: []string{"attr-board"},
		},
		{
			Id: &testpb.DutAttribute_Id{
				Value: "attr-design",
			},
			Aliases: []string{"attr-model"},
		},
		{
			Id: &testpb.DutAttribute_Id{
				Value: "swarming-pool",
			},
		},
		{
			Id: &testpb.DutAttribute_Id{
				Value: "misc-license",
			},
			Aliases: []string{"label-license"},
		},
	},
}

var boardPriorityList = &testplans.BoardPriorityList{
	BoardPriorities: []*testplans.BoardPriority{
		{
			SkylabBoard: "boardA", Priority: -100,
		},
		{
			SkylabBoard: "boardB", Priority: 100,
		},
	},
}

func TestToCTP1(t *testing.T) {
	ctx := context.Background()

	req := &testplans.GenerateTestPlanRequest{
		BuildbucketProtos: getSerializedBuilds(t),
	}

	resp, err := compatibility.ToCTP1(ctx,
		rand.New(rand.NewSource(7)),
		hwTestPlans, vmTestPlans, req, dutAttributeList, boardPriorityList, builderConfigs,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedResp := &testplans.GenerateTestPlanResponse{
		HwTestUnits: []*testplans.HwTestUnit{
			{
				Common: &testplans.TestUnitCommon{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "boardA",
					},
					BuilderName: "cq-builderA",
					BuildPayload: &testplans.BuildPayload{
						ArtifactsGsBucket: "testgsbucket",
						ArtifactsGsPath:   "testgspathA",
						FilesByArtifact: newStruct(t, map[string]interface{}{
							"IMAGE_ZIP": []interface{}{"file1", "file2"},
						}),
					},
				},
				HwTestCfg: &testplans.HwTestCfg{
					HwTest: []*testplans.HwTestCfg_HwTest{
						{
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-builderA.hw.suite1",
								Critical:    wrapperspb.Bool(true),
							},
							Suite:       "suite1",
							SkylabBoard: "boardA",
							Pool:        "DUT_POOL_QUOTA",
						},
						{
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-builderA.hw.suite2",
								Critical:    wrapperspb.Bool(true),
							},
							Suite:       "suite2",
							SkylabBoard: "boardA",
							Pool:        "DUT_POOL_QUOTA",
						},
						{
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-builderA.hw.suite4-multi-dut",
								Critical:    wrapperspb.Bool(false),
							},
							Suite: "suite4-multi-dut",
							TagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{`"group:somegroup"`},
								TagExcludes: []string{"informational"},
							},
							TotalShards: 1,
							SkylabBoard: "boardA",
							Pool:        "DUT_POOL_MULTI_DUT",
							Companions: []*testplans.TestCompanion{
								{
									Board: "boardCompanionA",
								},
								{
									Board: "pixelA",
									Config: &testpb.CompanionConfig{
										Config: &testpb.CompanionConfig_Android_{
											Android: &testpb.CompanionConfig_Android{
												GmsCorePackage: "latest-stable",
											},
										},
									},
								},
							},
						},
						{
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-builderA.model1.hw.suite3",
								Critical:    wrapperspb.Bool(false),
							},
							Suite: "suite3",
							TagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
								Tags:        []string{`"group:somegroup"`},
								TagExcludes: []string{"informational"},
							},
							TotalShards: 5,
							SkylabBoard: "boardA",
							SkylabModel: "model1",
							Licenses: []lab.LicenseType{
								lab.LicenseType_LICENSE_TYPE_WINDOWS_10_PRO,
								lab.LicenseType_LICENSE_TYPE_MS_OFFICE_STANDARD,
							},
							Pool: "DUT_POOL_QUOTA",
						},
					},
				},
			},
			{
				Common: &testplans.TestUnitCommon{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "boardA-kernelnext",
					},
					BuilderName: "cq-builderA-kernelnext",
					BuildPayload: &testplans.BuildPayload{
						ArtifactsGsBucket: "testgsbucket",
						ArtifactsGsPath:   "testgspathA-kernelnext",
						FilesByArtifact: newStruct(t, map[string]interface{}{
							"IMAGE_ZIP": []interface{}{"file1", "file2"},
						}),
					},
				},
				HwTestCfg: &testplans.HwTestCfg{
					HwTest: []*testplans.HwTestCfg_HwTest{
						{
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-builderA-kernelnext.model1.hw.suite-with-board-variant",
								Critical:    wrapperspb.Bool(false),
							},
							Suite:       "suite-with-board-variant",
							SkylabBoard: "boardA",
							SkylabModel: "model1",
							Pool:        "DUT_POOL_QUOTA",
							RunViaCft:   true,
						},
					},
				},
			},
			{
				Common: &testplans.TestUnitCommon{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "boardA",
					},
					BuilderName: "cq-builderA-asan",
					BuildPayload: &testplans.BuildPayload{
						ArtifactsGsBucket: "testgsbucket",
						ArtifactsGsPath:   "testgspathA-asan",
						FilesByArtifact: newStruct(t, map[string]interface{}{
							"IMAGE_ZIP": []interface{}{"file1", "file2"},
						}),
					},
				},
				HwTestCfg: &testplans.HwTestCfg{
					HwTest: []*testplans.HwTestCfg_HwTest{
						{
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-builderA-asan.hw.asan-suite",
								Critical:    wrapperspb.Bool(true),
							},
							Suite:                  "asan-suite",
							SkylabBoard:            "boardA",
							Pool:                   "DUT_POOL_QUOTA",
							RunViaCft:              true,
							EnableAutotestSharding: true,
						},
					},
				},
			},
		},
		DirectTastVmTestUnits: []*testplans.TastVmTestUnit{
			{
				Common: &testplans.TestUnitCommon{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "vmboardA",
					},
					BuilderName: "cq-vmBuilderA",
					BuildPayload: &testplans.BuildPayload{
						ArtifactsGsBucket: "testgsbucket",
						ArtifactsGsPath:   "testgspathA",
						FilesByArtifact: newStruct(t, map[string]interface{}{
							"IMAGE_ZIP": []interface{}{"file1", "file2"},
						}),
					},
				},
				TastVmTestCfg: &testplans.TastVmTestCfg{
					TastVmTest: []*testplans.TastVmTestCfg_TastVmTest{
						{
							SuiteName: "tast_vm_hwsec_cq",
							TastTestExpr: []*testplans.TastVmTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"dep:depB\"&&!\"informational\"&&(\"name:hwsec.*\"||\"name:cryptohome.*\")&&!\"name:firmware.*\"&&!\"name:arc.*\")",
								},
							},
							Common: &testplans.TestSuiteCommon{DisplayName: "cq-vmBuilderA.tast_vm.tast_vm_hwsec_cq", Critical: wrapperspb.Bool(true)},
						},
						{
							SuiteName: "tast_vm_suite1",
							TastTestExpr: []*testplans.TastVmTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"dep:depA\"&&!\"informational\")",
								},
							},
							Common: &testplans.TestSuiteCommon{DisplayName: "cq-vmBuilderA.tast_vm.tast_vm_suite1", Critical: wrapperspb.Bool(true)},
						},
					},
				},
			},
			{
				Common: &testplans.TestUnitCommon{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "vmboardA-arc-r",
					},
					BuilderName: "cq-vmBuilderA-arc-r",
					BuildPayload: &testplans.BuildPayload{
						ArtifactsGsBucket: "testgsbucket",
						ArtifactsGsPath:   "testgspathA-arc-r",
						FilesByArtifact: newStruct(t, map[string]interface{}{
							"IMAGE_ZIP": []interface{}{"file1", "file2"},
						}),
					},
				},
				TastVmTestCfg: &testplans.TastVmTestCfg{
					TastVmTest: []*testplans.TastVmTestCfg_TastVmTest{
						{
							SuiteName: "tast_vm_suite1",
							TastTestExpr: []*testplans.TastVmTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"dep:depA\"&&!\"informational\")",
								},
							},
							TastTestShard: &testplans.TastTestShard{
								TotalShards: 3,
								ShardIndex:  0,
							},
							Common: &testplans.TestSuiteCommon{DisplayName: "cq-vmBuilderA-arc-r.tast_vm.tast_vm_suite1_shard_1_of_3", Critical: wrapperspb.Bool(true)},
						},
						{
							SuiteName: "tast_vm_suite1",
							TastTestExpr: []*testplans.TastVmTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"dep:depA\"&&!\"informational\")",
								},
							},
							TastTestShard: &testplans.TastTestShard{
								TotalShards: 3,
								ShardIndex:  1,
							},
							Common: &testplans.TestSuiteCommon{DisplayName: "cq-vmBuilderA-arc-r.tast_vm.tast_vm_suite1_shard_2_of_3", Critical: wrapperspb.Bool(true)},
						},
						{
							SuiteName: "tast_vm_suite1",
							TastTestExpr: []*testplans.TastVmTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"dep:depA\"&&!\"informational\")",
								},
							},
							TastTestShard: &testplans.TastTestShard{
								TotalShards: 3,
								ShardIndex:  2,
							},
							Common: &testplans.TestSuiteCommon{DisplayName: "cq-vmBuilderA-arc-r.tast_vm.tast_vm_suite1_shard_3_of_3", Critical: wrapperspb.Bool(true)},
						},
					},
				},
			},
		},
		TastGceTestUnits: []*testplans.TastGceTestUnit{
			{
				Common: &testplans.TestUnitCommon{
					BuildTarget: &chromiumos.BuildTarget{
						Name: "vmboardA",
					},
					BuilderName: "cq-vmBuilderA",
					BuildPayload: &testplans.BuildPayload{
						ArtifactsGsBucket: "testgsbucket",
						ArtifactsGsPath:   "testgspathA",
						FilesByArtifact: newStruct(t, map[string]interface{}{
							"IMAGE_ZIP": []interface{}{"file1", "file2"},
						}),
					},
				},
				TastGceTestCfg: &testplans.TastGceTestCfg{
					TastGceTest: []*testplans.TastGceTestCfg_TastGceTest{
						{
							SuiteName: "tast_gce_suite2",
							GceMetadata: &testplans.TastGceTestCfg_TastGceTest_GceMetadata{
								Project:     "chromeos-gce-tests",
								Zone:        "us-central1-a",
								MachineType: "n1-standard-4",
								Network:     "chromeos-gce-tests",
								Subnet:      "us-central1",
							},
							TastTestExpr: []*testplans.TastGceTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"informational\")",
								},
							},
							TastTestShard: &testplans.TastTestShard{
								TotalShards: 2,
								ShardIndex:  0,
							},
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-vmBuilderA.tast_gce.tast_gce_suite2_shard_1_of_2",
								Critical:    wrapperspb.Bool(true),
							},
						},
						{
							SuiteName: "tast_gce_suite2",
							GceMetadata: &testplans.TastGceTestCfg_TastGceTest_GceMetadata{
								Project:     "chromeos-gce-tests",
								Zone:        "us-central1-a",
								MachineType: "n1-standard-4",
								Network:     "chromeos-gce-tests",
								Subnet:      "us-central1",
							},
							TastTestExpr: []*testplans.TastGceTestCfg_TastTestExpr{
								{
									TestExpr: "(\"group:mainline\"&&\"informational\")",
								},
							},
							TastTestShard: &testplans.TastTestShard{
								TotalShards: 2,
								ShardIndex:  1,
							},
							Common: &testplans.TestSuiteCommon{
								DisplayName: "cq-vmBuilderA.tast_gce.tast_gce_suite2_shard_2_of_2",
								Critical:    wrapperspb.Bool(true),
							},
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(expectedResp, resp, protocmp.Transform()); diff != "" {
		t.Errorf("ToCTP1Response returned unexpected diff (-want +got):\n%s", diff)
	}
}

func TestToCTP1Errors(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name             string
		hwTestPlans      []*test_api_v1.HWTestPlan
		vmTestPlans      []*test_api_v1.VMTestPlan
		dutAttributeList *testpb.DutAttributeList
		errRegexp        string
	}{
		{
			name:        "missing program DUT attribute",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "DutCriteria must contain at least one \"attr-program\" attribute",
		},
		{
			name:        "missing pool DUT attribute",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        `only DutCriteria with exactly one \"swarming-pool\" attribute are supported, got \[\]`,
		},
		{
			name:        "criteria with no values",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "only DutCriterion with at least one value supported",
		},
		{
			name:        "invalid DUT attribute",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-design",
											},
											Values: []string{"model1"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "fp",
											},
											Values: []string{"fp1"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "criterion .+ doesn't match any valid attributes",
		},
		{
			name:        "multiple pool values",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"testpoolA", "testpoolB"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"boardA", "boardB"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        `only DutCriteria with exactly one \"swarming-pool\" attribute are supported, got \[\"testpoolA\" \"testpoolB\"\]`,
		},
		{

			name:        "multiple design values",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"testpoolA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"boardA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-design",
											},
											Values: []string{"model1", "model2"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "only DutCriteria with one \"attr-design\" attribute are supported",
		},
		{
			name: "invalid DutAttributeList",
			dutAttributeList: &testpb.DutAttributeList{
				DutAttributes: []*testpb.DutAttribute{
					{
						Id: &testpb.DutAttribute_Id{
							Value: "otherdutattr",
						},
					},
				},
			},
			errRegexp: "\"attr-program\" not found in DutAttributeList",
		},
		{
			name:             "multiple programs with design",
			dutAttributeList: dutAttributeList,
			vmTestPlans:      vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					Id: &test_api_v1.HWTestPlan_TestPlanId{
						Value: "testplan1",
					},
					CoverageRules: []*testpb.CoverageRule{
						{
							Name: "invalidrule",
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA", "programB"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-design",
											},
											Values: []string{"model1"},
										},
									},
								},
							},
						},
					},
				},
			},
			errRegexp: "if \"attr-design\" is specified, multiple \"attr-programs\" cannot be used",
		},
		{
			name: "HW test without name",
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
								},
							},
							TestSuites: []*testpb.TestSuite{
								{
									Name: "",
									Spec: &testpb.TestSuite_TestCaseTagCriteria_{
										TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
											Tags: []string{"group:mainline"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "TestSuites must still specify a name if they are using TagCriteria",
		},
		{
			name:        "invalid VM test name",
			hwTestPlans: hwTestPlans,
			vmTestPlans: []*test_api_v1.VMTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							TestSuites: []*testpb.TestSuite{
								{
									Name: "vmsuite1",
									Spec: &testpb.TestSuite_TestCaseTagCriteria_{
										TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
											Tags: []string{"tagA"},
										},
									},
								},
							},
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA", "programB"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "VM suite names must start with either \"tast_vm\" or \"tast_gce\" in CTP1 compatibility mode",
		},
		{

			name:        "VM test with id list",
			hwTestPlans: hwTestPlans,
			vmTestPlans: []*test_api_v1.VMTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							TestSuites: []*testpb.TestSuite{
								{
									Name: "tast_vm_suite1",
									Spec: &testpb.TestSuite_TestCaseIds{
										TestCaseIds: &testpb.TestCaseIdList{
											TestCaseIds: []*testpb.TestCase_Id{
												{
													Value: "testcaseA",
												},
											},
										},
									},
								},
							},
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA", "programB"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "TestCaseIdLists are only valid for HW tests",
		},
		{
			name: "board variant with multiple programs",
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA", "programB"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
									ProvisionConfig: &testpb.ProvisionConfig{
										BoardVariant: "kernelnext",
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        `board_variant \(\"kernelnext\"\) and profile \(\"\"\) cannot be specified if multiple programs \(\[\"programA\" \"programB\"\]\) are specified`,
		},
		{
			name: "profile with multiple programs",
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"programA", "programB"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
									ProvisionConfig: &testpb.ProvisionConfig{
										Profile: "asan",
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        `board_variant \(\"\"\) and profile \(\"asan\"\) cannot be specified if multiple programs \(\[\"programA\" \"programB\"\]\) are specified`,
		},
		{
			name:        "invalid license attribute",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"board-A"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "misc-license",
											},
											Values: []string{"InvalidLicense"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "invalid LicenseType \".+\"",
		},
		{

			name:        "multiple license values",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"board-A"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "misc-license",
											},
											Values: []string{"LICENSE_TYPE_WINDOWS_10_PRO", "LICENSE_TYPE_MS_OFFICE_STANDARD"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "only exactly one value can be specified in \"misc-licence\" DutCriteria",
		},
		{
			name:        "program attribute specified twice",
			vmTestPlans: vmTestPlans,
			hwTestPlans: []*test_api_v1.HWTestPlan{
				{
					CoverageRules: []*testpb.CoverageRule{
						{
							DutTargets: []*testpb.DutTarget{
								{
									Criteria: []*testpb.DutCriterion{
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"board-A"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "attr-program",
											},
											Values: []string{"board-B"},
										},
										{
											AttributeId: &testpb.DutAttribute_Id{
												Value: "swarming-pool",
											},
											Values: []string{"DUT_POOL_QUOTA"},
										},
									},
								},
							},
						},
					},
				},
			},
			dutAttributeList: dutAttributeList,
			errRegexp:        "DutAttribute .+ specified twice",
		},
	}

	req := &testplans.GenerateTestPlanRequest{
		BuildbucketProtos: getSerializedBuilds(t),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := compatibility.ToCTP1(ctx,
				rand.New(rand.NewSource(7)),
				tc.hwTestPlans, tc.vmTestPlans, req, tc.dutAttributeList, boardPriorityList, builderConfigs,
			)
			if err == nil {
				t.Fatal("Expected error from ToCTP1")
			}

			matched, reErr := regexp.Match(tc.errRegexp, []byte(err.Error()))
			if reErr != nil {
				t.Fatal(reErr)
			}

			if !matched {
				t.Errorf("Expected error to match regexp %q, got %q", tc.errRegexp, err.Error())
			}
		})
	}
}
