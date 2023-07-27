# Copyright 2021 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

GOTESTFLAGS ?= -count 20

.PHONY: default
default: build

.PHONY:
build:
	@go vet ./...
	@go test $(GOTESTFLAGS) ./...
	go build ./...

.PHONY:
test-slow:
	go test ./... $(GOTESTFLAGS) -p 1

.PHONY: gen
gen:
	# Generate proto files
	@find . -name '*.pb.go' -delete
	go generate ./...
	# Generate config tree files
	RECOVERY_GENERATE_CONFIG_TREE=ON go test -run TestConfigTreeChanges

.PHONY:
test:
	go test ./...
