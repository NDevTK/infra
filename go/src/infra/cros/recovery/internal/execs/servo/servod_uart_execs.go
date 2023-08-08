// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"path/filepath"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/execs"
)

func startUartCaptureExec(ctx context.Context, info *execs.ExecInfo) error {
	err := servo.StartUartCapture(ctx, info.NewServod())
	return errors.Annotate(err, "start UART captiring").Err()
}
func stoptUartCaptureExec(ctx context.Context, info *execs.ExecInfo) error {
	err := servo.StopUartCapture(ctx, info.NewServod())
	return errors.Annotate(err, "stop UART captiring").Err()
}

func saveUartCaptureExec(ctx context.Context, info *execs.ExecInfo) error {
	dir := filepath.Join(info.GetLogRoot(), info.GetChromeos().GetServo().GetName())
	err := servo.SaveUartStreamToFiles(ctx, info.NewServod(), dir)
	return errors.Annotate(err, "save UART captiring").Err()
}

func init() {
	execs.Register("servod_start_uart_capture", startUartCaptureExec)
	execs.Register("servod_stop_uart_capture", stoptUartCaptureExec)
	execs.Register("servod_save_uart_capture", saveUartCaptureExec)
}
