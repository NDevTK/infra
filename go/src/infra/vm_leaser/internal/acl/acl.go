// Copyright 2023 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package acl

import (
	"go.chromium.org/luci/server/auth/rpcacl"
)

const (
	// TODO(justinsuen): Temporarily setting the auth group to CFS. Change to VM
	// Lab group when it is known.
	VMLabGroup = "mdb/chrome-fleet-software-team"
)

var RPCAccessInterceptor = rpcacl.Interceptor(rpcacl.Map{
	// Service metadata accessed through GRPC reflection should be accessible
	// only to authenticated users.
	"/grpc.reflection.v1alpha.ServerReflection/*": VMLabGroup,

	// Using the VM Leaser service requires the user or service to be part of the
	// VM Lab group.
	"/vm_leaser.api.v1.VMLeaserService/*": VMLabGroup,
})
