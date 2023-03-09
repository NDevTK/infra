#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

go build -o ./ ./cmd/buf ./cmd/protoc-gen-buf-breaking ./cmd/protoc-gen-buf-lint

cp ./buf                     "${PREFIX}"
cp ./protoc-gen-buf-breaking "${PREFIX}"
cp ./protoc-gen-buf-lint     "${PREFIX}"
