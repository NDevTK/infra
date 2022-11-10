# Copyright 2022 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
APP_PATH := $(PROJECT_PATH)
GCP_PROJECT := chrome-fleet-vm-leaser-cr-exp
GCP_REGION := us-west1

define help_message
Helpers for managing VM Leaser project in $(PROJECT_PATH):
  gen: regenerate all go bindings
  build: build the VM Leaser service image
  dev: run the VM Leaser service locally through Docker
  deploy: deploy the latest image to the ${GCP_PROJECT} project
endef

.PHONY: help
help:
	$(info $(help_message))

# Generate bindings
.PHONY: gen
gen:
	go generate ./...

# Build service image
.PHONY: build
build:
	CGO_ENABLED=0 go build -o vm-leaser-service $(PROJECT_PATH)/server
	docker build --tag=gcr.io/${GCP_PROJECT}/vm-leaser-service:latest --file=./Dockerfile .

# Run service locally
.PHONY: dev
dev: build
	docker run --rm --interactive --tty \
	--publish=50051:50051 \
	--env=PORT=50051 \
	gcr.io/${GCP_PROJECT}/vm-leaser-service:latest

# Deploy service
.PHONY: deploy
deploy: build
	gcloud run deploy vm-leaser-service \
	--image=gcr.io/${GCP_PROJECT}/vm-leaser-service:latest \
	--platform=managed \
	--use-http2 \
	--project=${GCP_PROJECT} \
	--region=${GCP_REGION}
