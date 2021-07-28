// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package logger provides an abstract representation of logging interfaces used by recovery lib.
package logger

import (
	"log"
)

// Logger represents a simple interface for logging data.
type Logger interface {
	// Debug log message at Debug level.
	Debug(format string, args ...interface{})
	// Info is like Debug, but logs at Info level.
	Info(format string, args ...interface{})
	// Warning is like Debug, but logs at Warning level.
	Warning(format string, args ...interface{})
	// Error is like Debug, but logs at Error level.
	Error(format string, args ...interface{})
	// Indenter provides access to logger intender.
	GetIndenter() Indenter
}

// NewLogger creates default logger.
func NewLogger() Logger {
	return &logger{}
}

// logger provides default implementation of Logger interface.
type logger struct {
	indenter Indenter
}

// Debug log message at Debug level.
func (l *logger) Debug(format string, args ...interface{}) {
	l.print(format, args...)
}

// Info is like Debug, but logs at Info level.
func (l *logger) Info(format string, args ...interface{}) {
	l.print(format, args...)
}

// Warning is like Debug, but logs at Warning level.
func (l *logger) Warning(format string, args ...interface{}) {
	l.print(format, args...)
}

// Error is like Debug, but logs at Error level.
func (l *logger) Error(format string, args ...interface{}) {
	l.print(format, args...)
}

// Indenter provides access to logger intender.
func (l *logger) GetIndenter() Indenter {
	if l.indenter == nil {
		l.indenter = NewIndenter()
	}
	return l.indenter
}

// Default logging logic for all levels.
func (l *logger) print(format string, args ...interface{}) {
	indent := GetIntentString(l.GetIndenter(), "\t")
	log.Printf(indent+format, args...)
}
