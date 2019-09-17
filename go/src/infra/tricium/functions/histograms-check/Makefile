# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

.PHONY: test clean

histograms-check:
	go build -o histograms-check

test: histograms-check
	./histograms-check -input=test -output=out

unit-test: 
	go test histograms-check_test.go histograms-check.go

clean:
	rm -rf out histograms-check

