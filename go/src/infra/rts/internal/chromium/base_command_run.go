// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chromium

import (
	"fmt"
	"os"

	"github.com/maruel/subcommands"
)

type BaseCommandRun struct {
	subcommands.CommandRunBase
}

func (r *BaseCommandRun) Done(err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
