# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

.PHONY: test
GO_WCT_PATH = ../../../../tools/wct

default: help

help:
	@echo "Available commands:"
	@sed -n '/^[a-zA-Z0-9_]*:/s/:.*//p' < Makefile

node_modules:
	npm install

build: node_modules
	polymer build

# TODO: Allow wct.go to work with "package paths"/"bare paths"
# that are used in Polymer 3 JS module imports.
# test: node_modules
# 	xvfb-run go run $(GO_WCT_PATH)/wct.go \
# 		$(GO_WCT_PATH)/monkeypatch.go \
# 		-base=. -prefix=test/ -chrome=google-chrome
#
# test_persist: node_modules
# 	go run $(GO_WCT_PATH)/wct.go \
# 		$(GO_WCT_PATH)/monkeypatch.go \
# 		-base=. -prefix=test/ -chrome=google-chrome -persist
