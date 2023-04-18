# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

.PHONY: test clean

inclusive:
	go build

test: inclusive
	./inclusive -input=test -output=out

clean:
	rm -rf out inclusive
