// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package vector

import (
	"testing"
)

func TestVectorCompare(t *testing.T) {
	t.Parallel()

	a := *New(1, 0, 0)
	b := *New(1, 0, 1)
	c := *New(0, 1, 1)
	d := *New(1, 0, 0)
	cases := []struct {
		A      Vector
		B      Vector
		Expect bool
	}{
		{a, a, false},
		{b, a, false},
		{a, b, true},
		{c, a, true},
		{d, a, false},
	}
	for _, expect := range cases {
		actual := expect.A.Less(expect.B)
		if actual != expect.Expect {
			t.Errorf("%+v < %+v = %+v, want %+v",
				expect.A, expect.B, actual, expect.Expect)
		}
	}
}
