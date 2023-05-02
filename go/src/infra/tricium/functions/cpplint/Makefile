# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

.PHONY: test clean

cpplint_parser: python3 update_cpplint.py
	go build -o cpplint_parser

test: cpplint_parser
	./cpplint_parser -input=test -output=out

python3:
	echo "infra/3pp/tools/cpython3/linux-amd64 version:2@3.8.10.chromium.28" | \
			cipd ensure -ensure-file - -root python3

update_cpplint.py:
	curl https://chromium.googlesource.com/chromium/tools/depot_tools.git/+/main/cpplint.py\?format\=TEXT \
		| base64 --decode > cpplint.py
	chmod +x cpplint.py

clean:
	rm -rf cpplint_parser cpplint.py python3 out
