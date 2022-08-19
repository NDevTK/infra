// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitiles

import (
	"context"
)

type MockedGitilesClient struct {
	Data map[string]string // Data for MockedGitilesClient to return
}

func (cl *MockedGitilesClient) sendRequest(c context.Context, url string, params map[string]string) (string, error) {
	if val, ok := cl.Data[url]; ok {
		return val, nil
	}
	return `{"logs":[]}`, nil
}

func MockedGitilesClientContext(c context.Context, data map[string]string) context.Context {
	return context.WithValue(c, MockedGitilesClientKey, &MockedGitilesClient{
		Data: data,
	})
}
