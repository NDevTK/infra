// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package scheduler

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("infra/qscheduler/qslib")
