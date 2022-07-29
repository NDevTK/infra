#!/usr/bin/env vpython

import os
import subprocess
import sys
if sys.platform == 'win32':
  import msvcrt

print("Hello")
vpython_exe = os.getenv("VPYTHON_TEST_EXE")

nb_child = 3
procs = [
    subprocess.Popen(
        [vpython_exe,
         "child%d/child.py" % ((nb_child % 2) + 1),
         str(i)]) for i in xrange(nb_child * 5)
]
for p in procs:
  p.wait()
  assert (p.returncode == 0)

with open(__file__) as f:
  fd = f.fileno()
  if sys.platform == 'win32':
    fd = msvcrt.get_osfhandle(fd)
  p = subprocess.Popen(
      [vpython_exe, "fdtest/child.py", str(fd)], close_fds=False)
  p.wait()
  assert (p.returncode == 0)
