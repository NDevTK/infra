// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cipkg

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"reflect"
	"strings"
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
	Platform string

	// The command of the execution. In most cases it's the executable binary.
	// The standard executor (builtins.Execute) provides some basic operations
	// under "builtin:" prefix, including builtin:fetchUrls, builtin:cipdEnsure
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

// Calculate a unique ID from the content of a derivation. The ID may also be
// used as file or directory name on the local filesystem.
func (d Derivation) ID() string {
	h := sha256.New()
	enc := base32.HexEncoding.WithPadding(base32.NoPadding)

	// sha256 shouldn't return any write error.
	writeField := func(name string, ss ...string) {
		if _, err := fmt.Fprint(h, name); err != nil {
			panic(err)
		}
		for _, s := range ss {
			if _, err := fmt.Fprintf(h, "\t%s", enc.EncodeToString([]byte(s))); err != nil {
				panic(err)
			}
		}
		if _, err := fmt.Fprintln(h); err != nil {
			panic(err)
		}
	}

	v := reflect.ValueOf(d)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			writeField(v.Type().Field(i).Name, f.String())
		case reflect.Slice:
			writeField(v.Type().Field(i).Name, f.Interface().([]string)...)
		default:
			panic("unexpected field type")
		}
	}

	// We want to keep the hash as short as possible to avoid reaching the path
	// length limit on windows.
	// Using base32 instead of base64 because filesystem is not promised to be
	// case-sensitive.
	return fmt.Sprintf("%s-%s", d.Name, strings.ToLower(enc.EncodeToString(h.Sum(nil)[:16])))
}
