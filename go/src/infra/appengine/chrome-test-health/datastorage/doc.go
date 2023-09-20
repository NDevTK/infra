// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package datastorage contains clients which are
// essentially wrappers over various data services such
// as google cloud datastore. The code in this package
// ensures that we isolate communication to
// these services from our business logic.
//
// This isolation will help in better unit testing
// as well as a single place to change this communication
// logic as opposed to directly changing the business logic.
package datastorage
