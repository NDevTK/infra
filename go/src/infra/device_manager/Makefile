# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
APP_PATH := $(PROJECT_PATH)
GCP_PROJECT := fleet-device-manager-dev
GCP_REGION := us-central1
LUCI_AUTH_SERVICE := chrome-infra-auth-dev.appspot.com

define help_message

Helpers for managing Device Manager project in $(PROJECT_PATH)

For the main Device Lease service:
  gen: regenerate all go bindings
  build: build the Device Lease service image
  dev: run the Device Lease service locally
  push-image: build and push the latest image to the ${GCP_PROJECT} project

endef

.PHONY: help
help:
	$(info $(help_message))

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
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o device-lease-service $(PROJECT_PATH)/cmd/device_lease_service

# Run service locally
.PHONY: dev
dev: build
	./device-lease-service \
	--auth-service-host=${LUCI_AUTH_SERVICE} \
	--grpc-addr=":50051" \
	--http-addr=":8800" \
	--admin-addr=":8900"

# Build Docker service
.PHONY: build-docker
build-docker: build
	docker build --tag=gcr.io/${GCP_PROJECT}/device-lease-service:latest --file=./cmd/device_lease_service/docker/Dockerfile .

# Push Docker image to GCP
.PHONY: push-image
push-image: build-docker
	docker push gcr.io/${GCP_PROJECT}/device-lease-service:latest

# Run Docker service locally
.PHONY: dev-docker
dev-docker: build-docker
	docker run --rm --interactive --tty --user $(id -u):$(id -g) \
	-p 50051:50051 \
	-p 8800:8800 \
	-p 8900:8900 \
	-v ~/.config/chrome_infra/auth:/tokens \
	gcr.io/${GCP_PROJECT}/device-lease-service:latest \
		-cloud-project ${GCP_PROJECT} \
		-auth-service-host=${LUCI_AUTH_SERVICE} \
		-grpc-addr :50051 \
		-http-addr :8800 \
		-admin-addr :8900 \
		-token-cache-dir /tokens
