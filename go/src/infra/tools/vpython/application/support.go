// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package application

import (
	"infra/tools/vpython/python"

	"github.com/luci/luci-go/common/errors"

	"golang.org/x/net/context"
)

var appKey = "infra/tools/vpython/application.A"

func withApplication(c context.Context, a *A) context.Context {
	return context.WithValue(c, &appKey, a)
}

func getApplication(c context.Context) *A {
	return c.Value(&appKey).(*A)
}

func run(c context.Context, fn func(context.Context) error) int {
	err := fn(c)

	switch t := errors.Unwrap(err).(type) {
	case nil:
		return 0

	case python.Error:
		return int(t)

	default:
		errors.Log(c, err)
		return 1
	}
}
