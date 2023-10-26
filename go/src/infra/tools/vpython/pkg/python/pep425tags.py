# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import os
import pip._internal.utils.compatibility_tags as compatibility_tags

tags = [{
    "python": t.interpreter,
    "abi": t.abi,
    "platform": t.platform
} for t in compatibility_tags.get_supported()]
with open(os.path.join(os.environ['out'], 'pep425tags.json'), 'w') as f:
  f.write(json.dumps(tags))
