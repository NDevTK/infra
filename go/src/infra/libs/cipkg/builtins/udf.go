// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builtins

import (
	"context"
	"os/exec"
	"strings"
)

// UDFBuilder simply runs any function registered based on the name. This allows
// users to use a go function as a builder with some caution. The function
// itself is not part of the derivation, which means the derivation won't
// change if the function's behaviour changed. So the user is responsible for
// changing the derivation to represent the possible output change if the
// UserDefinedFunction changed.
const UDFBuilderPrefix = BuiltinBuilderPrefix + "udf:"

type UserDefinedFunction func(ctx context.Context, cmd *exec.Cmd) error

var udfs map[string]UserDefinedFunction = make(map[string]UserDefinedFunction)

// Register a function which will be executed by "builtin:udf:${name}" builder.
// DO NOT USE THIS unless you understand what you are doing.
// Derivation is not able to detect any behaviour change of the function.
// You MUST add and maintain your own version string to indicate a behaviour
// change of the function.
func RegisterUserDefinedFunction(name string, f UserDefinedFunction) {
	if _, ok := udfs[name]; ok {
		panic("duplicated user defined function")
	}
	udfs[name] = f
}

func executeUserdefinedFunction(ctx context.Context, cmd *exec.Cmd) error {
	return udfs[strings.TrimPrefix(cmd.Path, UDFBuilderPrefix)](ctx, cmd)
}
