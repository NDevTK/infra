package main

import (
	"context"
	"infra/cros/internal/assert"
	"testing"

	"github.com/google/go-cmp/cmp"
	testpb "go.chromium.org/chromiumos/config/go/test/api"
	tpv2 "go.chromium.org/chromiumos/infra/proto/go/test_platform/v2"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestRunOrch(t *testing.T) {
	ctx := context.Background()
	request := &tpv2.RequestBeta{
		Request: &tpv2.RequestBeta_HwTestRequest{
			HwTestRequest: &tpv2.HWTestRequestBeta{
				TestSpecs: []*tpv2.HWTestRequestBeta_TestSpec{
					{
						Rules: &testpb.CoverageRule{
							Name: "test_rule1",
							TestSuites: []*testpb.TestSuite{
								{
									Name: "test_suite1",
									Spec: &testpb.TestSuite_TestCaseTagCriteria_{
										TestCaseTagCriteria: &testpb.TestSuite_TestCaseTagCriteria{
											Tags: []string{"kernel"},
										},
									},
								},
								{
									Name: "test_suite2",
									Spec: &testpb.TestSuite_TestCaseIds{
										TestCaseIds: &testpb.TestCaseIdList{
											TestCaseIds: []*testpb.TestCase_Id{
												{
													Value: "suiteA",
												},
											},
										},
									},
								},
							},
							DutCriteria: []*testpb.DutCriterion{
								{
									AttributeId: &testpb.DutAttribute_Id{
										Value: "dutattr1",
									},
									Values: []string{"valA", "valB"},
								},
							},
						},
					},
				},
			},
		},
	}

	err := RunOrch(ctx, request)
	assert.NilError(t, err)
}

func TestRunOrchErrors(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		request      *tpv2.RequestBeta
		errorMessage string
	}{
		{
			"empty request",
			&tpv2.RequestBeta{},
			"at least one TestSpec in request required",
		},
		{
			"empty CoverageRule",
			&tpv2.RequestBeta{
				Request: &tpv2.RequestBeta_HwTestRequest{
					HwTestRequest: &tpv2.HWTestRequestBeta{
						TestSpecs: []*tpv2.HWTestRequestBeta_TestSpec{
							{
								Rules: &testpb.CoverageRule{},
							},
						},
					},
				},
			},
			"at least one DutCriterion required in each CoverageRule",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RunOrch(ctx, tc.request)
			assert.ErrorContains(t, err, tc.errorMessage)
		})
	}
}

func TestGetRequestedDimensions(t *testing.T) {
	ctx := context.Background()

	dutCriteria := []*testpb.DutCriterion{
		{
			AttributeId: &testpb.DutAttribute_Id{
				Value: "dutattr1",
			},
			Values: []string{"valA", "valB"},
		}, {
			AttributeId: &testpb.DutAttribute_Id{
				Value: "dutattr2",
			},
			Values: []string{"valC"},
		},
	}

	expectedDims := []*bbpb.RequestedDimension{
		{
			Key:   "dutattr1",
			Value: "valA",
		},
		{
			Key:   "dutattr2",
			Value: "valC",
		},
	}

	dims, err := GetRequestedDimensions(ctx, dutCriteria)

	assert.NilError(t, err)

	if diff := cmp.Diff(expectedDims, dims, protocmp.Transform()); diff != "" {
		t.Errorf("GetRequestedDimensions(%s) returned unexpected diff (-want +got):\n%s", dutCriteria, diff)
	}
}

func TestGetRequestedDimensionsErrors(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		dutCriteria  []*testpb.DutCriterion
		errorMessage string
	}{
		{
			"no id",
			[]*testpb.DutCriterion{
				{
					Values: []string{"valA"},
				},
			},
			"DutAttribute id must be set",
		},
		{
			"no values",
			[]*testpb.DutCriterion{
				{
					AttributeId: &testpb.DutAttribute_Id{
						Value: "dutattr1",
					},
					Values: []string{},
				},
			},
			"at least one value must be set on DutAttributes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetRequestedDimensions(ctx, tc.dutCriteria)
			assert.ErrorContains(t, err, tc.errorMessage)
		})
	}
}
