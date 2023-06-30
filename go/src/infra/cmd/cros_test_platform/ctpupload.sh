#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Original at go/ctp-cipd-script

set -eu

# User defined vars
# Ex: "/usr/local/google/home/azrahman/chromiumos/infra/infra/go/src/infra/cmd/"
root_dir=""
# Ex: "/usr/local/google/home/azrahman/chromiumos/infra/infra/go/bin/"
bin_dir=""
# Ex: "/usr/local/google/home/azrahman/chromiumos/infra/infra/build/packages"
pkgs_dir=""
# Update this to whatever reference name you would like to set
ref_name="$USER-test"

# System generated vars
luciexe_pkg="luciexe"
ctp_pkg="cros_test_platform"
ctp_dir="$root_dir/$ctp_pkg"
luciexe_dir="$ctp_dir/$luciexe_pkg"

build_and_copy_bin() {
 # Expects args (1)root_dir, (2)bin_dir, (3)pkg_name
 echo " ----- building $3..."
 echo "building binary..."
 cd $1
 export CGO_ENABLED=0
 go build

 echo "Changing permission..."
 sudo chmod a+rwx $3

 echo "Copy binary..."
 sudo cp $3 $2
}

upload_and_set_ref() {
 # Expects args (1)pkgs_dir, (2)pkg_name, (3)ref_name
 echo " ----- Uploading $2..."
 echo "Changing dir..."
 cd $1

 echo "Create CIPD package..."
 cipd create -pkg-def $2.yaml -pkg-var exe_suffix: -verbose

 echo " "
 echo "Read version:"
 read -r -p "Enter version: " ver

 echo "Set tag..."
 cipd set-ref chromiumos/infra/$2/linux-amd64 -ref=$3 -version=$ver
}

# Main code
# Build luciexe
build_and_copy_bin $luciexe_dir $bin_dir $luciexe_pkg
# Build ctp
build_and_copy_bin $ctp_dir $bin_dir $ctp_pkg
# Upload binary and set ref
upload_and_set_ref $pkgs_dir $ctp_pkg $ref_name