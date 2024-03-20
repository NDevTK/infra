// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package jobs contains the handlers registered and called by Cloud Scheduler.
//
// The HTTP endpoints exposed by this module perform necessary authorization
// checks and route requests to registered handlers, collecting monitoring
// metrics from them.
//
// Note that you still need to configure Cloud Scheduler jobs. By default,
// registered handlers are exposed as "/internal/cron/<handler-id>" endpoints.
// This URL path should be used when configuring Cloud Scheduler jobs.
package jobs
