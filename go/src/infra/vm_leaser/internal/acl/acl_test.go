// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package acl

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRPCAccessInterceptor(t *testing.T) {
	t.Parallel()

	interceptor := RPCAccessInterceptor.Unary()

	check := func(ctx context.Context, service, method string) codes.Code {
		info := &grpc.UnaryServerInfo{
			FullMethod: fmt.Sprintf("/%s/%s", service, method),
		}
		_, err := interceptor(ctx, nil, info, func(context.Context, interface{}) (interface{}, error) {
			return nil, nil
		})
		return status.Code(err)
	}

	Convey("Anonymous", t, func() {
		ctx := auth.WithState(context.Background(), &authtest.FakeState{})

		So(check(ctx, "unknown.API", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "grpc.reflection.v1alpha.ServerReflection", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "grpc.reflection.v1.ServerReflection", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "chromiumos.test.api.VMLeaserService", "Something"), ShouldEqual, codes.PermissionDenied)
	})

	Convey("Authenticated, but not authorized", t, func() {
		ctx := auth.WithState(context.Background(), &authtest.FakeState{
			Identity:       "user:someone@example.com",
			IdentityGroups: []string{"some-random-group"},
		})

		So(check(ctx, "unknown.API", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "grpc.reflection.v1alpha.ServerReflection", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "grpc.reflection.v1.ServerReflection", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "chromiumos.test.api.VMLeaserService", "Something"), ShouldEqual, codes.PermissionDenied)
	})

	Convey("Authorized", t, func() {
		ctx := auth.WithState(context.Background(), &authtest.FakeState{
			Identity:       "user:someone@example.com",
			IdentityGroups: []string{VMLabGroup},
		})

		So(check(ctx, "unknown.API", "Something"), ShouldEqual, codes.PermissionDenied)
		So(check(ctx, "grpc.reflection.v1alpha.ServerReflection", "Something"), ShouldEqual, codes.OK)
		So(check(ctx, "grpc.reflection.v1.ServerReflection", "Something"), ShouldEqual, codes.OK)
		So(check(ctx, "chromiumos.test.api.VMLeaserService", "Something"), ShouldEqual, codes.OK)
	})
}
