// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/execs/wifirouter/ssh/mocks"
	"infra/cros/recovery/tlw"
)

func Test_remoteFileContentsMatch(t *testing.T) {
	type args struct {
		remoteFilePath       string
		matchRegex           string
		expectFileExistsCall bool
		mockFileExistsResult bool
		expectCatFileCall    bool
		mockCatFileResult    string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"file does not exist",
			args{
				"some_file.txt",
				"",
				true,
				false,
				false,
				"",
			},
			false,
			false,
		},
		{
			"bad regex",
			args{
				"some_file.txt",
				"???",
				true,
				true,
				false,
				"",
			},
			false,
			true,
		},
		{
			"empty file matches simple regex",
			args{
				"some_file.txt",
				".*",
				true,
				true,
				true,
				"",
			},
			true,
			false,
		},
		{
			"gale match regex",
			args{
				lsbReleaseFilePath,
				lsbReleaseMatchIfGale,
				true,
				true,
				true,
				"SOME_VALUE=\nCHROMEOS_RELEASE_BOARD=gale\nOTHER_VALUE=asd\n",
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRunner := mocks.NewMockRunner(ctrl)
			if tt.args.expectFileExistsCall {
				var fileExistsRunResult ssh.RunResult
				if tt.args.mockFileExistsResult {
					fileExistsRunResult = &tlw.RunResult{
						ExitCode: 0,
					}
				} else {
					fileExistsRunResult = &tlw.RunResult{
						ExitCode: 1,
					}
				}
				mockRunner.EXPECT().
					RunForResult(gomock.Any(), gomock.Any(), false, "test", "-f", gomock.Eq(tt.args.remoteFilePath)).
					Return(fileExistsRunResult)
			}
			if tt.args.expectCatFileCall {
				mockRunner.EXPECT().
					Run(gomock.Any(), gomock.Any(), "cat", gomock.Eq(tt.args.remoteFilePath)).
					Return(tt.args.mockCatFileResult, nil)
			}
			got, err := RemoteFileContentsMatch(context.Background(), mockRunner, tt.args.remoteFilePath, tt.args.matchRegex)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteFileContentsMatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RemoteFileContentsMatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replaceInvalidWifiRouterFeatures(t *testing.T) {
	tests := []struct {
		name            string
		featuresInitial []labapi.WifiRouterFeature
		featuresAfter   []labapi.WifiRouterFeature
	}{
		{
			"empty list",
			[]labapi.WifiRouterFeature{},
			[]labapi.WifiRouterFeature{},
		},
		{
			"nil list",
			nil,
			nil,
		},
		{
			"all valid",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
			},
		},
		{
			"single invalid",
			[]labapi.WifiRouterFeature{
				-1,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
		},
		{
			"mixed valid and invalid",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
				1234124,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				-1,
				-254,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replaceInvalidWifiRouterFeatures(tt.featuresInitial)
			if !reflect.DeepEqual(tt.featuresInitial, tt.featuresAfter) {
				t.Errorf("replaceInvalidWifiRouterFeatures() got = %v, want %v", tt.featuresInitial, tt.featuresAfter)
			}
		})
	}
}

func Test_removeDuplicateWifiRouterFeatures(t *testing.T) {
	tests := []struct {
		name        string
		featuresArg []labapi.WifiRouterFeature
		want        []labapi.WifiRouterFeature
	}{
		{
			"empty list",
			[]labapi.WifiRouterFeature{},
			nil,
		},
		{
			"nil list",
			nil,
			nil,
		},
		{
			"no duplicates",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
		},
		{
			"with duplicates",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
		},
		{
			"with unknown value duplicate",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature(1235679),
				labapi.WifiRouterFeature(1235679),
				labapi.WifiRouterFeature(652325),
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature(1235679),
				labapi.WifiRouterFeature(652325),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeDuplicateWifiRouterFeatures(tt.featuresArg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeDuplicateWifiRouterFeatures() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SortWifiRouterFeaturesByName(t *testing.T) {
	tests := []struct {
		name            string
		featuresInitial []labapi.WifiRouterFeature
		featuresAfter   []labapi.WifiRouterFeature
	}{
		{
			"empty list",
			[]labapi.WifiRouterFeature{},
			[]labapi.WifiRouterFeature{},
		},
		{
			"nil list",
			nil,
			nil,
		},
		{
			"already sorted",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
		},
		{
			"named sort",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
			},
		},
		{
			"unknown name value sort",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature(99901),
				labapi.WifiRouterFeature(99902),
				labapi.WifiRouterFeature(99900),
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature(99900),
				labapi.WifiRouterFeature(99901),
				labapi.WifiRouterFeature(99902),
			},
		},
		{
			"mixed sort",
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature(99900),
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
				labapi.WifiRouterFeature(99901),
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
				labapi.WifiRouterFeature(99902),
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
				labapi.WifiRouterFeature(99900),
				labapi.WifiRouterFeature(99901),
				labapi.WifiRouterFeature(99902),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortWifiRouterFeaturesByName(tt.featuresInitial)
			if !reflect.DeepEqual(tt.featuresInitial, tt.featuresAfter) {
				t.Errorf("SortWifiRouterFeaturesByName() got = %v, want %v", tt.featuresInitial, tt.featuresAfter)
			}
		})
	}
}

func TestCollectCommonWifiRouterFeatures(t *testing.T) {
	type args struct {
		featureSets [][]labapi.WifiRouterFeature
	}
	tests := []struct {
		name string
		args args
		want []labapi.WifiRouterFeature
	}{
		{
			"nil feature sets",
			args{
				nil,
			},
			nil,
		},
		{
			"no feature sets",
			args{
				[][]labapi.WifiRouterFeature{},
			},
			nil,
		},
		{
			"single empty feature set",
			args{
				[][]labapi.WifiRouterFeature{
					{},
				},
			},
			[]labapi.WifiRouterFeature{},
		},
		{
			"single non-empty feature set",
			args{
				[][]labapi.WifiRouterFeature{
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
					},
				},
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
			},
		},
		{
			"two feature sets, second is empty",
			args{
				[][]labapi.WifiRouterFeature{
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
					},
					{},
				},
			},
			[]labapi.WifiRouterFeature{},
		},
		{
			"two feature sets with no common features",
			args{
				[][]labapi.WifiRouterFeature{
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_BE,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
					},
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_B,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_G,
					},
				},
			},
			[]labapi.WifiRouterFeature{},
		},
		{
			"two feature sets with some common features",
			args{
				[][]labapi.WifiRouterFeature{
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_BE,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
					},
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_B,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_G,
					},
				},
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
			},
		},
		{
			"three feature sets with some common features",
			args{
				[][]labapi.WifiRouterFeature{
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_BE,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_N,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
					},
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_N,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_B,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AX_E,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_G,
					},
					{
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_B,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_N,
						labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_G,
					},
				},
			},
			[]labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_N,
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectCommonWifiRouterFeatures(tt.args.featureSets)
			SortWifiRouterFeaturesByName(got)
			SortWifiRouterFeaturesByName(tt.want)
			if !(len(tt.want) == 0 && len(got) == 0) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectCommonWifiRouterFeatures() = %v, want %v", got, tt.want)
			}
		})
	}
}
