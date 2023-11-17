#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -eux

# We want to import TestPlan protos from a source that should be mirrored in
# `go.chromium.org/chromiumos/infra/proto/src`
# https://chromium.googlesource.com/chromiumos/infra/proto/+/main/src/test_platform/request.proto

# To download that code and all of it's dependencies we need to capture them in
# the proto path flag in the protoc command. Any import or dependency of that
# import needs to be found within one of the proto paths.

# The ../../../../ paths should correspond to locations where our dependencies
# on chromeos protos reside.

protoc \
--proto_path=. \
--proto_path=../../../../../go.chromium.org/chromiumos/config/proto \
--proto_path=../../../../../go.chromium.org/chromiumos/infra/proto/src \
--go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
*.proto