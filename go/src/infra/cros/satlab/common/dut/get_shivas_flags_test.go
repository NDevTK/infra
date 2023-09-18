// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"reflect"
	"testing"
)

func TestMakeGetShivasFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		getCmd *GetDUT
		want   flagmap
	}{
		{
			name:   "default",
			getCmd: &GetDUT{},
			want: flagmap{
				"namespace": []string{"os"},
				"json":      []string{},
			},
		},
		{
			name: "all fields",
			getCmd: &GetDUT{
				Zones:         []string{"input_zone"},
				Racks:         []string{"input_racks"},
				Machines:      []string{"input_machines"},
				Prototypes:    []string{"input_prototypes"},
				Servos:        []string{"input_servos"},
				Servotypes:    []string{"input_servotypes"},
				Switches:      []string{"input_switches"},
				Rpms:          []string{"input_rpms"},
				Pools:         []string{"input_pools"},
				HostInfoStore: true,
				Namespace:     "os-partner",
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
				"json":            []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeGetDUTShivasFlags(tt.getCmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeGetShivasFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
