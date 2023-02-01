// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package site contains site local constants for the shivas
package site

import (
	"testing"

	ufsUtil "infra/unifiedfleet/app/util"
)

func TestEnvFlags_Namespace(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		validNSList []string
		defaultNS   string
		want        string
		wantErr     bool
	}{
		{
			"nil namespace",
			"",
			AllNamespaces,
			"",
			"",
			true,
		},
		{
			"non-client, non-valid namespace",
			"fake",
			AllNamespaces,
			"",
			"fake",
			true,
		},
		{
			"client, non-valid namespace",
			ufsUtil.OSPartnerNamespace,
			[]string{ufsUtil.OSNamespace},
			"",
			ufsUtil.OSPartnerNamespace,
			true,
		},
		{
			"non-client, valid namespace",
			"fake",
			[]string{"fake", ufsUtil.OSNamespace},
			"",
			"fake",
			true,
		},
		{
			"happy path",
			ufsUtil.OSPartnerNamespace,
			OSLikeNamespaces,
			"",
			ufsUtil.OSPartnerNamespace,
			false,
		},
		{
			"happy path ignores default",
			ufsUtil.OSPartnerNamespace,
			OSLikeNamespaces,
			"default",
			ufsUtil.OSPartnerNamespace,
			false,
		},
		{
			"no input with default",
			"",
			AllNamespaces,
			ufsUtil.OSPartnerNamespace,
			ufsUtil.OSPartnerNamespace,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := EnvFlags{
				namespace: tt.namespace,
			}
			got, err := f.Namespace(tt.validNSList, tt.defaultNS)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvFlags.Namespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EnvFlags.Namespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
