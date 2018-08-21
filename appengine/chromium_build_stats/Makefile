# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

generate: gopath/src/ninjalog/ninjalog.proto
	go generate ./...

deploy:
	./goenv.sh goapp deploy --application chromium-build-stats-staging --version $(shell git log -1 --pretty=format:git-%H) default

build:
	cd default; ../goenv.sh goapp build

test:
	cd default; ../goenv.sh goapp test
