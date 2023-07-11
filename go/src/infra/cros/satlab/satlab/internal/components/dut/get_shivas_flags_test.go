// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"infra/cros/satlab/satlab/internal/site"
	"reflect"
	"testing"
)

func TestMakeGetShivasFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		getCmd *getDUT
		want   flagmap
	}{
		{
			name:   "default",
			getCmd: &getDUT{},
			want: flagmap{
				"namespace": []string{"os"},
			},
		},
		{
			name: "all fields",
			getCmd: &getDUT{
				shivasGetDUT{
					zones:             []string{"input_zone"},
					racks:             []string{"input_racks"},
					machines:          []string{"input_machines"},
					prototypes:        []string{"input_prototypes"},
					servos:            []string{"input_servos"},
					servotypes:        []string{"input_servotypes"},
					switches:          []string{"input_switches"},
					rpms:              []string{"input_rpms"},
					pools:             []string{"input_pools"},
					wantHostInfoStore: true,
					envFlags:          site.MakeEnvFlagsForTesting("os-partner"),
				},
			},
			want: flagmap{
				"zone":            []string{"input_zone"},
				"rack":            []string{"input_racks"},
				"machine":         []string{"input_machines"},
				"prototype":       []string{"input_prototypes"},
				"servo":           []string{"input_servos"},
				"servotype":       []string{"input_servotypes"},
				"switch":          []string{"input_switches"},
				"rpms":            []string{"input_rpms"},
				"pools":           []string{"input_pools"},
				"host-info-store": []string{},
				"namespace":       []string{"os-partner"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeGetShivasFlags(tt.getCmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeGetShivasFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
