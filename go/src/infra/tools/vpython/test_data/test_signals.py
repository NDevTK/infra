# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import signal
import sys
import threading
import time

# File descriptor #3 is ExtraFiles[0], our signal write pipe.
signalW = os.fdopen(3, 'w')

# Wait for interrupt signal.
signalled = False
def signal_handler(sig, frame):
  global signalled
  print 'Received SIGINT!'
  signal.signal(signal.SIGINT, signal.SIG_DFL)
  signalled = True
signal.signal(signal.SIGINT, signal_handler)

# Notify our parent that it's ready to send the signal.
signalW.close()

# Loop indefinitely. Our parent process is responsible for killing us.
print 'Waiting for signal...'
while not signalled:
  time.sleep(.1)
print 'Exiting after confirming signal.'
