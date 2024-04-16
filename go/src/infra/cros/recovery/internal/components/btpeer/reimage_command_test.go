// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"strings"
	"testing"
	"time"

	"infra/cros/recovery/internal/components"
)

// mockResult is a runner that always returns the same value to test result parsing.
func mockResult(output []string) components.Runner {
	return func(context.Context, time.Duration, string, ...string) (string, error) {
		return strings.Join(output, "\n"), nil
	}
}

func TestGetPartitionInfo(t *testing.T) {
	tests := []struct {
		name           string
		output         []string
		wantErr        bool
		skipFree       bool
		expectedResult []*partitionInfo
	}{
		{
			name: "pass_skip_free",
			output: []string{
				"BYT",
				"/dev/mmcblk0:100%:sd/mmc:512:512:msdos:SD",
				"1:0.01%:0.85%:0.84%:fat32::lba",
				"2:0.85%:11.9%:11.0%:ext4::",
				"1:11.9%:100%:88.1%:free",
			},
			wantErr:  false,
			skipFree: true,
			expectedResult: []*partitionInfo{
				{
					Number: 1,
					Start:  0.01,
					End:    0.85,
					Size:   0.85 - 0.01,
					Type:   "fat32",
				},
				{
					Number: 2,
					Start:  0.85,
					End:    11.9,
					Size:   11.9 - 0.85,
					Type:   "ext4",
				},
			},
		},
		{
			name: "pass_with_free",
			output: []string{
				"BYT",
				"/dev/mmcblk0:100%:sd/mmc:512:512:msdos:SD",
				"1:0.01%:0.85%:0.84%:fat32::lba",
				"2:0.85%:11.9%:11.0%:ext4::",
				"1:11.9%:100%:88.1%:free",
			},
			wantErr:  false,
			skipFree: false,
			expectedResult: []*partitionInfo{
				{
					Number: 1,
					Start:  0.01,
					End:    0.85,
					Size:   0.85 - 0.01,
					Type:   "fat32",
				},
				{
					Number: 2,
					Start:  0.85,
					End:    11.9,
					Size:   11.9 - 0.85,
					Type:   "ext4",
				},
				{
					Number: 1,
					Start:  11.9,
					End:    100,
					Size:   100 - 11.9,
					Type:   "free",
				},
			},
		},
		{
			name: "pass_empty",
			output: []string{
				"BYT",
				"/dev/mmcblk0:100%:sd/mmc:512:512:msdos:SD",
				"1:0.1%:100%:88.1%:free",
			},
			wantErr:  false,
			skipFree: false,
			expectedResult: []*partitionInfo{
				{
					Number: 1,
					Start:  0.1,
					End:    100,
					Size:   100 - 0.1,
					Type:   "free",
				},
			},
		},
		{
			name: "fail_malformed_start",
			output: []string{
				"1:0..1%:100%:88.1%:free",
			},
			wantErr:        true,
			skipFree:       false,
			expectedResult: nil,
		},
		{
			name: "fail_malformed_start",
			output: []string{
				"1:0..1%:100%:88.1%:free",
			},
			wantErr:        true,
			skipFree:       false,
			expectedResult: nil,
		},
		{
			name: "fail_malformed_end",
			output: []string{
				"1:0.1%:10..0%:88.1%:free",
			},
			wantErr:        true,
			skipFree:       false,
			expectedResult: nil,
		},
		{
			name: "fail_negative_size",
			output: []string{
				"1:10%:1%:88.1%:free",
			},
			wantErr:        true,
			skipFree:       false,
			expectedResult: nil,
		},
		{
			name: "fail_negative_start",
			output: []string{
				"-1.0:10%:100%:88.1%:free",
			},
			wantErr:        true,
			skipFree:       false,
			expectedResult: nil,
		},
		{
			name: "fail_negative_end",
			output: []string{
				"1:1%:-10%:88.1%:free",
			},
			wantErr:        true,
			skipFree:       false,
			expectedResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &partitionHelper{
				device: "/dev/mmcblk0",
				runner: mockResult(test.output),
			}
			info, err := p.GetPartitionInfo(context.Background(), "%", test.skipFree)
			if len(info) != len(test.expectedResult) {
				t.Errorf("TestPartitionInfo: invalid result, got %d results expected %d", len(info), len(test.expectedResult))
				return
			}
			for i := range test.expectedResult {
				if *info[i] != *test.expectedResult[i] {
					t.Errorf("TestPartitionInfo: invalid result, got %v expected %v", info[i], test.expectedResult[i])
					return
				}
			}
			if (err != nil) != test.wantErr {
				t.Errorf("TestPartitionInfo: error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}
