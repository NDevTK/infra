// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

// LocalVMLeaserServiceEndpoint is the local endpoint for VM Leaser.
const LocalVMLeaserServiceEndpoint = "127.0.0.1"

// LocalVMLeaserServiceEndpoint is the port to connect to local endpoint.
const LocalVMLeaserServicePort = 50051

// StagingVMLeaserServiceEndpoint is the staging cloud project for VM Leaser.
const StagingVMLeaserServiceEndpoint = "staging.vmleaser.api.cr.dev"

// StagingVMLeaserServiceEndpoint is the port to connect to staging endpoint.
const StagingVMLeaserServicePort = 443

// TODO(b/266128274): Replace endpoint with the deployed prod endpoint.
// ProdVMLeaserServiceEndpoint is the prod cloud project for VM Leaser.
const ProdVMLeaserServiceEndpoint = "staging.vmleaser.api.cr.dev"

// ProdVMLeaserServiceEndpoint is the port to connect to prod endpoint.
const ProdVMLeaserServicePort = 443
