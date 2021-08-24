// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package logger

// helpers_test.go contains utilities for creating fake versions of various
// types for use in tests. For example, it contains a fake logger, which can
// be used in tests to ensure that specific log messages were written at
// specific severity levels.

// logCB is a callback that intercepts an attempt to write a log message.
// This is used inside the fake logger to test what log messages were emitted.
type logCB = func(level string, format string, args []interface{})

// FakeLogger is an implementation of the logger interface
// that is suitable for use in tests. It records calls as
// necessary to support tests.
type fakeLogger struct {
	logCallback logCB
}

// NewFakeLogger creates a new logger that's suitable for
// use in tests.
func newFakeLogger(logCallback logCB) Logger {
	return &fakeLogger{
		logCallback: logCallback,
	}
}

// Debug intercepts a debug-level message.
func (l *fakeLogger) Debug(format string, args ...interface{}) {
	l.logCallback("debug", format, args)
}

// Info intercepts an info-level message.
func (l *fakeLogger) Info(format string, args ...interface{}) {
	l.logCallback("info", format, args)
}

// Warning intercepts a warning-level message.
func (l *fakeLogger) Warning(format string, args ...interface{}) {
	l.logCallback("warning", format, args)
}

// Error intercepts an error-level message.
func (l *fakeLogger) Error(format string, args ...interface{}) {
	l.logCallback("error", format, args)
}

// IndentLogging intercepts an indent attempt.
func (l *fakeLogger) IndentLogging() {
	// Intentionally do nothing.
}

// DedentLogging intercepts a dedent attempt.
func (l *fakeLogger) DedentLogging() {
	// Intentionally do nothing.
}
