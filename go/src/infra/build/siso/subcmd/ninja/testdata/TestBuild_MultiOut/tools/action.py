# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import sys

for out in sys.argv[1:]:
  with open(out, "w") as f:
    f.write("")
