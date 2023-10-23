// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package acl

import (
	"go.chromium.org/luci/server/auth/rpcacl"
)

const (
	// This is a LUCI Auth group set in both the dev and prod LUCI Auth Services.
	VMLabGroup = "vm-leaser-access"
)

var RPCAccessInterceptor = rpcacl.Interceptor(rpcacl.Map{
	// Service metadata accessed through GRPC reflection should be accessible
	// only to authenticated users.
	"/grpc.reflection.v1alpha.ServerReflection/*": VMLabGroup,
	"/grpc.reflection.v1.ServerReflection/*":      VMLabGroup,

	// Using the VM Leaser service requires the user or service to be part of the
	// VM Lab group.
	"/chromiumos.test.api.VMLeaserService/*": VMLabGroup,
})
