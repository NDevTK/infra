// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"log"
	"testing"
)

func TestConfigTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		data     interface{}
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
			log.Printf("generated conf for %q is:\n%s", tc.name, got)
		})
	}
}
