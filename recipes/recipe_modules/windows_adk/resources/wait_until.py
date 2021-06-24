#!/usr/bin/python

import sys
import os
import time

f = open(sys.argv[1], 'r')
while True:
  for l in f.readlines():
    if sys.argv[2] not in l:
      continue
    else:
      exit(0)
  time.sleep(60)
