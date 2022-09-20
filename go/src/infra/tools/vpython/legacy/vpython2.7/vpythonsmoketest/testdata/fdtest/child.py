#!/usr/bin/env vpython

# This test verifies that a file handle passed from the parent process can be
# opened by the child.

import os
import sys
if sys.platform == 'win32':
  import msvcrt

fd = int(sys.argv[1])
print('fd=%d' % fd)
if sys.platform == 'win32':
  fd = msvcrt.open_osfhandle(fd, os.O_RDONLY)
f = os.fdopen(fd, 'r')
f.close()
