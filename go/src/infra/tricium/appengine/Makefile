# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Commands here assume that the Go environment is set up
# by running `eval $(../../../../env.py)`.

default: help

help:
	@echo "Available commands:"
	@sed -n '/^[a-zA-Z0-9_]*:/s/:.*//p' < Makefile

deploy_prod: build
	gae.py upload --switch --app-id=tricium-prod

deploy_dev: build
	gae.py upload --switch --app-id=tricium-dev

build:
	cd frontend/ui && make build
