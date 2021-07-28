// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package logger

// Indenter provides holder to manage indent of logs.
type Indenter interface {
	Increment()
	Decrement()
	GetIndent() int
}

// NewIndenter creates default implementation of intender.
func NewIndenter() Indenter {
	return &indenter{
		indent: 0,
	}
}

// indenter local representations of LogIndenter interface.
type indenter struct {
	indent int
}

// Increment increments the indent.
func (i *indenter) Increment() {
	i.indent += 1
}

// Decrement decrements the indent.
func (i *indenter) Decrement() {
	if i.indent > 0 {
		i.indent -= 1
	}
}

// GetIndent provides the value of indent.
func (i *indenter) GetIndent() int {
	return i.indent
}

// Generate indent string before messages.
// Default indent is tab (`\t`).
func GetIntentString(i Indenter, indentStr string) string {
	if i == nil || i.GetIndent() == 0 {
		return ""
	}
	if indentStr == "" {
		indentStr = "\t"
	}
	is := []byte(indentStr)
	count := i.GetIndent()
	b := make([]byte, count*len(is))
	for i := 0; i < count; i++ {
		b = append(b, is...)
	}
	return string(b)
}
