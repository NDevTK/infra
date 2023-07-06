// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package ssh is a collection of ssh utilities for executing remote ssh
// commands on hosts.
//
// TODO(jaredbennett/otabek): Move this out from under the wifirouter execs to a general location
//
// Some common, simple bash command functions are included in this package as
// well, but more complex or uncommon bash command functions should not be
// included in this package. Rather, those should be closer to their usage and
// have unit tests.
package ssh
