#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

cd cmd/protoc-gen-go-grpc
go build -o "${PREFIX}/protoc-gen-go-grpc" google.golang.org/grpc/cmd/protoc-gen-go-grpc
