// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cookflags

import "fmt"

func dumpStr(to *[]string, name, val string, dflt ...string) {
	checkVal := ""
	if len(dflt) > 0 {
		checkVal = dflt[0]
	}
	if val != checkVal {
		*to = append(*to, "-"+name, val)
	}
}

func dumpList(to *[]string, name string, vals []string) {
	arg := "-" + name
	for _, v := range vals {
		*to = append(*to, arg, v)
	}
}

func dumpMap(to *[]string, name string, vals map[string]string) {
	arg := "-" + name
	for k, v := range vals {
		if v == "" {
			*to = append(*to, arg, k)
		} else {
			*to = append(*to, arg, fmt.Sprintf("%s=%s", k, v))
		}
	}
}

func dumpBool(to *[]string, name string, val bool) {
	if val {
		*to = append(*to, "-"+name)
	}
}
