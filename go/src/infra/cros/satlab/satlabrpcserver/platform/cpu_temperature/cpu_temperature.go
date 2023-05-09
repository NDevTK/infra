// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package cpu_temperature

// ICPUTemperature is the interface that provides what functions we can
// use. The main idea is for different platform to implement their own logic.
type ICPUTemperature interface {
	// GetCurrentCPUTemperature Get the current temperature of CPU
	GetCurrentCPUTemperature() (float32, error)
}
