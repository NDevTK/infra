# Copyright 2022 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
APP_PATH := $(PROJECT_PATH)
GCP_PROJECT := chrome-fleet-vm-leaser-cr-exp
GCP_REGION := us-west1

define help_message

Helpers for managing VM Leaser project in $(PROJECT_PATH)

For the main VM Leaser service:
  gen: regenerate all go bindings
  build: build the VM Leaser service image
  dev: run the VM Leaser service locally
  push-image: build and push the latest image to the ${GCP_PROJECT} project

For the cron service:
  build-cron: build the VM Leaser cron image
  dev-cron: run the VM Leaser cron service locally

endef

.PHONY: help
help:
	$(info $(help_message))

# Generate bindings
.PHONY: gen
gen:
	go generate ./...

# Run Go tests
.PHONY: test
test:
	go test ./... -coverprofile=coverage.out

# Check code coverage
.PHONY: cover
cover:
	go tool cover -func=coverage.out

# Build service image
.PHONY: build
build:
	CGO_ENABLED=0 go build -o vm-leaser-service $(PROJECT_PATH)/cmd/vm_leaser

# Run service locally
.PHONY: dev
dev: build
	./vm-leaser-service --auth-service-host="chrome-infra-auth-dev.appspot.com"

# Build Docker service
.PHONY: build-docker
build-docker: build
	docker build --tag=gcr.io/${GCP_PROJECT}/vm-leaser-service:latest --file=./cmd/vm_leaser/docker/Dockerfile .

# Push Docker image to GCP
.PHONY: push-image
push-image: build-docker
  docker push gcr.io/${GCP_PROJECT}/vm-leaser-service:latest

# Run Docker service locally
.PHONY: dev-docker
dev-docker: build-docker
	docker run --rm --interactive --tty \
	--publish=50051:50051 \
	--env=PORT=50051 \
	gcr.io/${GCP_PROJECT}/vm-leaser-service:latest

# Deploy Docker service
.PHONY: deploy-docker
deploy-docker: build-docker
	gcloud run deploy vm-leaser-service \
	--image=gcr.io/${GCP_PROJECT}/vm-leaser-service:latest \
	--platform=managed \
	--use-http2 \
	--project=${GCP_PROJECT} \
	--region=${GCP_REGION}

# Build cron service
.PHONY: build-cron
build-cron:
	go build -o cron-service $(PROJECT_PATH)/cmd/cron

# Run cron service locally
.PHONY: dev-cron
dev-cron: build-cron
	./cron-service \
	--http-addr="127.0.0.1:8081" \
	--admin-addr="127.0.0.1:8901" \
	--auth-service-host="chrome-infra-auth-dev.appspot.com"
