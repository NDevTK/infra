// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cipkg

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

// Derivation is the atomic level of a build step. It should contain all
// information used during the execution in its definition. Because Derivation
// is use its content to determine if the output will be different, it should
// maintain best-effort reproducibility to keep result consistent.
// NOTE: ${out} is not part of the derivation. We can't determine the output
// directory before we have a deterministic derivation so it has to be added
// during the execution.
type Derivation struct {
	// The name of this derivation. It may be used to refer the derivation in
	// definition.
	Name string

	// The platform where this derivation will be executed.
	// TODO: Specify the format? (e.g amd64_linux)
	Platform string

	// The command of the execution. In most cases it's the executable binary.
	// The standard executor (builtins.Execute) provides some basic operations
	// under "builtin:" prefix, including builtin:fetchUrl, builtin:cipdEnsure
	// and others. In most cases builtin commands should be used with their own
	// generator (e.g. builtins.CIPDEnsure).
	Builder string

	// Arguments passed to the builder.
	Args []string

	// Environments for the execution.
	Env []string

	// The IDs of all derivations referred by this derivation.
	Inputs []string
}

// Calculate a unique ID from the content of a derivation
func (d Derivation) ID() string {
	h := sha256.New()
	gob.NewEncoder(h).Encode(d)
	return fmt.Sprintf("%s-%x", d.Name, h.Sum(nil))
}
