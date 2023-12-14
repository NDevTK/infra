// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func fullyPopulated(m proto.Message) bool {
	r := m.ProtoReflect()
	rd := m.ProtoReflect().Descriptor()
	for fields, i := rd.Fields(), 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if f.HasPresence() {
			fv := r.Get(f)
			if !fv.IsValid() {
				return false
			}
			if f.Kind() == protoreflect.MessageKind {
				if !fullyPopulated(fv.Message().Interface()) {
					return false
				}
			}
		}
	}
	return true
}
