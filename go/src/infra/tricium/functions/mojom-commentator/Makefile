# Copyright 2019 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

.PHONY: test clean

mojom-commentator:
	go build

test: mojom-commentator
	./mojom-commentator -input=test -output=out

clean:
	rm -rf out mojom-commentator
