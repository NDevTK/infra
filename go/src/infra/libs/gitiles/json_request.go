// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitiles

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type jsonResult struct {
	data interface{}
	err  error
}

type jsonRequest struct {
	urlPath    string
	fact       typeFactory
	resultChan chan<- jsonResult
}

type typeFactory func() interface{}

func (j jsonRequest) URLPath() string { return j.urlPath + "?format=JSON" }
func (j jsonRequest) Method() string  { return "GET" }
func (j jsonRequest) Body() io.Reader { return nil }

func (j jsonRequest) Process(resp *http.Response, err error) {
	data, err := j.process(resp, err)
	j.resultChan <- jsonResult{data, err}
}

func (j jsonRequest) process(resp *http.Response, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}

	bufreader := bufio.NewReader(resp.Body)
	firstFour, err := bufreader.Peek(4)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(firstFour, []byte(")]}'")) {
		bufreader.Read(make([]byte, 4))
	}

	data := j.fact()
	err = json.NewDecoder(bufreader).Decode(data)

	if err != nil {
		return nil, err
	}
	return data, nil
}
