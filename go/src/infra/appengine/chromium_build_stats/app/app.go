// Copyright 2014 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package app implements chromium-build-stats.appspot.com services.
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/context/ctxhttp"
)

const (
	topHTML = `
<html>
<head><title>chromium-build-stats</title></head>
<body>
<h1>chromium-build-stats</h1>
<form action="/">
<label for="loc">compile step URL or gs URI:</label><input type="text" name="loc" />
<input type="submit" value="submit"><input type="reset">
</form>

<hr />
See <a href="https://docs.google.com/a/chromium.org/document/d/16TdPTIIZbtAarXZIMJdiT9CePG5WYCrdxm5u9UuHXNY/edit?pli=1#heading=h.xgjl2srtytjt">design doc</a>
</body>
</html>
`
)

var ninjaLogRE = regexp.MustCompile(`gs://chrome-goma-log/.*/ninja_log.*\.gz`)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, req *http.Request) {
	loc := req.FormValue("loc")
	if loc != "" {
		if strings.HasPrefix(loc, "http://build.chromium.org/") {
			ctx := req.Context()
			resp, err := ctxhttp.Get(ctx, nil, loc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				http.Error(w, fmt.Sprintf("%s reply %s", loc, resp.Status), resp.StatusCode)
				return
			}
			s := bufio.NewScanner(resp.Body)
			for s.Scan() {
				if m := ninjaLogRE.Find(s.Bytes()); m != nil {
					loc = string(m)
					break
				}
			}
			if err := s.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if strings.HasPrefix(loc, "gs://chrome-goma-log") {
			logPath := strings.TrimPrefix(loc, "gs://chrome-goma-log")
			basename := path.Base(loc)
			if strings.HasPrefix(basename, "ninja_log.") {
				http.Redirect(w, req, "/ninja_log"+logPath, http.StatusSeeOther)
				return
			}
		}
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, topHTML)
}
