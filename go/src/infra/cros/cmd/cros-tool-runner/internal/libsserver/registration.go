// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package common defines shared resources across registration and test libs service.
package libsserver

import (
	"errors"
	"strings"
)

// LibReg represents the information provided by a single library
// when registering.
type LibReg struct {
	Name        string   `json:"name"`
	Registry    string   `json:"registry"`
	Version     int      `json:"version"`
	Image       string   `json:"image"`
	ExecCmd     []string `json:"exec_cmd"`
	LogDir      string   `json:"log_dir"`
	Port        string   `json:"port"`
	ServoPort   string   `json:"servo_port"`
	Ping        string   `json:"ping"`
	APIType     string   `json:"api_type"`
	Owners      []string `json:"owners"`
	Description string   `json:"description"`
}

// Validate returns an error if the registration info has any issues.
func (r *LibReg) Validate() error {
	problems := []string{}
	if r.Name == "" {
		problems = append(problems, "Name cannot be blank")
	}
	if r.APIType == "" {
		problems = append(problems, "APIType cannot be blank")
	}
	if r.APIType != "REST" {
		problems = append(problems, "Unrecognized API type of "+r.APIType)
	}
	if r.Image == "" {
		problems = append(problems, "Image name cannot be blank")
	}
	if len(r.Owners) == 0 {
		problems = append(problems, "Provide at least one owner")
	}
	if r.Description == "" {
		problems = append(problems, "Description string cannot be blank")
	}
	if len(problems) != 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}
