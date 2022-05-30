#!/bin/bash
# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Load our installation utility functions.
. /install-util.sh

# Install missing packages for system Python modules.
if [ -x /usr/bin/apt-get ]; then
  apt-get install -y zlib1g-dev libbz2-dev libltdl-dev texi2html texinfo python
  apt-get clean --yes
elif [ -x /usr/bin/yum ]; then
  yum install -y zlib-devel bzip2-devel ncurses-devel sqlite-devel texi2html texinfo
  yum clean all
else
  echo "UKNOWN package platform."
  exit 1
fi

# The CentOS images also don't have `nproc`, so add it.
if ! which nproc; then
  echo '#!/bin/bash' > /usr/bin/nproc
  echo 'grep processor < /proc/cpuinfo | wc -l' >> /usr/bin/nproc
  chmod +x /usr/bin/nproc
fi
