# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

.PHONY: mallet
mallet:
	@$(RM) -rf ./mallet
	env CGO_ENABLED=0 go build -o ./mallet -- .

.PHONY: test
test:
	go test ./...

.PHONY: check
check: test

.PHONY: clean
clean:
	$(RM) ./mallet

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...
