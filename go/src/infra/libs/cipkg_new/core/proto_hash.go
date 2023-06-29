// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package core

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

const anyName = "google.protobuf.Any"

// Generate an n-character stable base32 encoded id from the proto message.
// Base32 should be safe for url, filesystem path and other common applications.
func StableID(m proto.Message, n uint) (string, error) {
	h := sha256.New()
	if err := StableHash(h, m); err != nil {
		return "", err
	}
	enc := base32.HexEncoding.WithPadding(base32.NoPadding)
	return strings.ToLower(enc.EncodeToString(h.Sum(nil)[:n])), nil
}

// StableHash returns a hash of protobuf message that is guaranteed to be
// stable over time.
// If the message contains any unknown field, StableHash will return an error.
//
// TODO(fancl): move this to go.chromium.org/luci/common/proto.
func StableHash(h hash.Hash, m proto.Message) error {
	return hashMessage(h, m.ProtoReflect())
}

func hashValue(h hash.Hash, v protoreflect.Value) error {
	switch v := v.Interface().(type) {
	case int32:
		hashNumber(h, uint64(v))
	case int64:
		hashNumber(h, uint64(v))
	case uint32:
		hashNumber(h, uint64(v))
	case uint64:
		hashNumber(h, v)
	case float32:
		hashNumber(h, uint64(math.Float32bits(v)))
	case float64:
		hashNumber(h, math.Float64bits(v))
	case string:
		hashNumber(h, uint64(len(v)))
		h.Write([]byte(v))
	case []byte:
		hashNumber(h, uint64(len(v)))
		h.Write(v)
	case protoreflect.EnumNumber:
		hashNumber(h, uint64(v))
	case protoreflect.Message:
		if err := hashMessage(h, v); err != nil {
			return err
		}
	case protoreflect.List:
		if err := hashList(h, v); err != nil {
			return err
		}
	case protoreflect.Map:
		if err := hashMap(h, v); err != nil {
			return err
		}
	case bool:
		var b uint64
		if v {
			b = 1
		}
		hashNumber(h, b)
	default:
		return fmt.Errorf("unknown type: %T", v)
	}
	return nil
}

func hashMessage(h hash.Hash, m protoreflect.Message) error {
	if m.Descriptor().FullName() == anyName {
		a, err := m.Interface().(*anypb.Any).UnmarshalNew()
		if err != nil {
			return err
		}
		return hashMessage(h, a.ProtoReflect())
	}

	if m.GetUnknown() != nil {
		return fmt.Errorf("unknown fields cannot be hashed")
	}

	// Collect a sorted list of populated message fields.
	var fds []protoreflect.FieldDescriptor
	m.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		fds = append(fds, fd)
		return true
	})
	sort.Slice(fds, func(i, j int) bool { return fds[i].Number() < fds[j].Number() })

	// Iterate over message fields.
	for _, fd := range fds {
		hashNumber(h, uint64(fd.Number()))
		if err := hashValue(h, m.Get(fd)); err != nil {
			return err
		}
	}
	return nil
}

func hashList(h hash.Hash, lv protoreflect.List) error {
	hashNumber(h, uint64(lv.Len()))
	for i := 0; i < lv.Len(); i++ {
		if err := hashValue(h, lv.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func hashMap(h hash.Hash, mv protoreflect.Map) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	hashNumber(h, uint64(mv.Len()))

	// Collect a sorted list of populated map entries.
	var ks []protoreflect.MapKey
	mv.Range(func(k protoreflect.MapKey, _ protoreflect.Value) bool {
		ks = append(ks, k)
		return true
	})
	sort.Slice(ks, func(i, j int) bool {
		ki, kj := ks[i], ks[j]
		switch ki.Interface().(type) {
		case bool:
			return !ki.Bool() && kj.Bool()
		case int32, int64:
			return ki.Int() < kj.Int()
		case uint32, uint64:
			return ki.Uint() < kj.Uint()
		case string:
			return ki.String() < kj.String()
		default:
			panic("invalid map key type")
		}
	})

	// Iterate over map entries.
	for _, k := range ks {
		if err := hashValue(h, k.Value()); err != nil {
			return err
		}
		if err := hashValue(h, mv.Get(k)); err != nil {
			return err
		}
	}
	return nil
}

func hashNumber(h hash.Hash, v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	h.Write(b[:])
}
