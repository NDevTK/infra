// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"strings"
	"testing"
)

func TestConfigTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		template  string
		data      interface{}
		wantLines []string
	}{
		{
			name:     "nginx",
			template: nginxTemplate,
			data: &nginxConf{
				CacheSizeInGB:     100,
				Port:              1234,
				L7Port:            4321,
				L7Servers:         []string{"1.1.1.1", "2.2.2.2"},
				OtelTraceEndpoint: "http://localhost:5678",
			},
			wantLines: []string{
				"server 1.1.1.1:4321;",
				"server 2.2.2.2:4321;",
				"listen *:1234",
				"max_size=100g",
			},
		},
		{
			name:     "keepalived",
			template: keepalivedTempalte,
			data: &keepalivedConf{
				ServiceIP:   "8.8.8.8",
				ServicePort: 9999,
				RealServers: []string{"1.1.1.1", "2.2.2.2"},
				Interface:   "eth0",
				LBAlgo:      "wlc",
			},
			wantLines: []string{
				"interface eth0",
				"8.8.8.8",
				"virtual_server 8.8.8.8 9999{",
				"real_server 1.1.1.1 9999{",
				"connect_port 9999",
				"real_server 2.2.2.2 9999{",
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := genConfig("nginx", tc.template, tc.data)
			if err != nil {
				t.Errorf("genConfig(%q) err %q, want nil", tc.name, err)
			}
			for _, w := range tc.wantLines {
				if !strings.Contains(got, w) {
					t.Errorf("genConfig(%q) got %q, doesn't contain %q", tc.name, got, w)
				}
			}
		})
	}
}
