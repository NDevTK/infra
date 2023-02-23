#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREFIX="$1"

cp -rf --parents * "$PREFIX"/

# Create a wrapper batch file in our expected bin/ install location.
mkdir -p "$PREFIX"/bin
cat <<EOF > "$PREFIX"/bin/perl.bat
@echo off
setlocal
set PATH=%~dp0..\perl\site\bin;%~dp0..\perl\bin;%~dp0c\bin;%PATH%
perl.exe %*
EOF
