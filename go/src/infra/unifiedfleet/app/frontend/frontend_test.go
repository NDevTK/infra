// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"testing"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"google.golang.org/grpc"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/util"
)

var nilHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func mockUser(ctx context.Context, user string) context.Context {
	return auth.WithState(ctx, &authtest.FakeState{
		Identity: identity.Identity(fmt.Sprintf("user:%s", user)),
	})
}

func mockGroupMembership(ctx context.Context, group string) context.Context {
	state := auth.GetState(ctx).(*authtest.FakeState)
	state.IdentityGroups = append(state.IdentityGroups, group)

	return ctx
}

func loadACLConfig(ctx context.Context) context.Context {
	alwaysUseACLConfig := config.Config{
		PartnerACLGroups: []string{"all-sfp-partners"},
	}

	return config.Use(ctx, &alwaysUseACLConfig)
}

func TestPartnerInterceptor(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		user     string
		ns       string
		sfpGroup bool
		wantErr  bool
	}{
		{
			name:     "googler not in SfP group",
			user:     "test@google.com",
			ns:       util.OSNamespace,
			sfpGroup: false,
			wantErr:  false,
		},
		{
			name:     "googler in SfP group",
			user:     "test@google.com",
			ns:       util.OSNamespace,
			sfpGroup: true,
			wantErr:  false,
		},
		{
			name:     "non-googler in SfP group",
			user:     "test@gmail.com",
			ns:       util.OSNamespace,
			sfpGroup: true,
			wantErr:  true,
		},
		// non-googler NOT in sfp group will not have access denied, as that
		// should be taken care of by not granting other ACLs like RPC-level
		// access.
		{
			name:     "non-googler not in SfP group",
			user:     "test@gmail.com",
			ns:       util.OSNamespace,
			sfpGroup: false,
			wantErr:  false,
		},
		{
			name:     "googler not in SfP group in partner ns",
			user:     "test@google.com",
			ns:       util.OSPartnerNamespace,
			sfpGroup: false,
			wantErr:  false,
		},
		{
			name:     "googler in SfP group in partner ns",
			user:     "test@google.com",
			ns:       util.OSPartnerNamespace,
			sfpGroup: true,
			wantErr:  false,
		},
		{
			name:     "non-googler in SfP group in partner ns",
			user:     "test@gmail.com",
			ns:       util.OSPartnerNamespace,
			sfpGroup: true,
			wantErr:  false,
		},
		{
			name:     "non-googler not in SfP group in partner ns",
			user:     "test@gmail.com",
			ns:       util.OSPartnerNamespace,
			sfpGroup: false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		ctx := context.Background()
		ctx = memory.Use(ctx)
		ctx = loadACLConfig(ctx)

		t.Run(tt.name, func(t *testing.T) {
			ctx, err := util.SetupDatastoreNamespace(ctx, tt.ns)
			if err != nil {
				t.Errorf("error setting up ns: %s", err)
			}
			ctx = mockUser(ctx, tt.user)
			if tt.sfpGroup {
				ctx = mockGroupMembership(ctx, "all-sfp-partners")
			}

			_, respErr := PartnerInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "test"}, nilHandler)
			if (respErr != nil) != tt.wantErr {
				t.Errorf("partnerInterceptor() error = %v, wantErr %v", respErr, tt.wantErr)
				return
			}
		})
	}
}
