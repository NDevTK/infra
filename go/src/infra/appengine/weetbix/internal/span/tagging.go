// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package span

import (
	"context"

	"go.chromium.org/luci/server/span"
	"google.golang.org/grpc"
)

// SpannerDefaultsInterceptor returns a gRPC interceptor that adds default
// Spanner request options to the context.
//
// The request tag will be set the to RPC method name.
//
// See also ModifyRequestOptions in luci/server/span.
func SpannerDefaultsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = span.ModifyRequestOptions(ctx, func(opts *span.RequestOptions) {
			opts.Tag = info.FullMethod
		})
		return handler(ctx, req)
	}
}
