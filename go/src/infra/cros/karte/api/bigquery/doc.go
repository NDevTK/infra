// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package kbqpb is the bigquery proto API of karte.
//
// These protos should implement the ValueSaver interface, which allows
// them to control the names of their fields.
//
// The protos also control the schema of the BigQuery tables belonging to Karte.
// These protos are intentionally very simple. Karte uses flat BigQuery tables in order to make ad hoc SQL queries more ergonomic.
package kbqpb
