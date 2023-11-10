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
			data:     &nginxConf{},
		},
		{
			name:     "keepalived",
			template: keepalivedTempalte,
			data:     &keepalivedConf{},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := genConfig("nginx", nginxTemplate, &nginxConf{})
			if err != nil {
				t.Errorf("genConfig(%q) err %q, want nil", tc.name, err)
			}
			log.Printf("generated conf for %q is:\n%s", tc.name, got)
		})
	}
}
