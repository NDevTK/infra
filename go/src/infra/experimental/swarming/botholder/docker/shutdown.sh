#!/bin/bash

# Notify the root container process (botholder) that Swarming wants to restart.
kill -s SIGUSR1 1
