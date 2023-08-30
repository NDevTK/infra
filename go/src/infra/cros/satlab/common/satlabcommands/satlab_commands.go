// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package satlabcommands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/executor"
)

// Decision is a classification of a line in a file.
// Lines may be kept, modified, or deleted.
// Functions that process lines of text are split conceptually
// into a decision which classifies lines and a transformation
// which only applies to selected lines.
type Decision int

const (
	Unknown Decision = iota
	Keep
	Reject
	// Modify is used only by replacing things.
	Modify
)

// GetHostIdentifier gets the host identifier value.
//
// Note that this command always returns the identifier in lowercase.
func GetDockerHostBoxIdentifier(executor executor.IExecCommander) (string, error) {
	fmt.Fprintf(os.Stderr, "Get host identifier: run %s\n", paths.GetHostIdentifierScript)
	out, err := executor.Exec(exec.Command(paths.GetHostIdentifierScript))
	// Immediately normalize the satlab prefix to lowercase. It will save a lot of
	// trouble later.
	return strings.ToLower(TrimOutput(out)), errors.Annotate(err, "get host identifier").Err()
}
