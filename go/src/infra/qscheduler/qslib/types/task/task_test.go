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

package task

import "testing"

func TestLabelSetEquals(t *testing.T) {
	type tuple struct {
		a      LabelSet
		b      LabelSet
		expect bool
	}
	a := LabelSet([]string{"label1"})
	b := LabelSet([]string{"label1"})
	c := LabelSet{}
	d := LabelSet{}
	e := LabelSet([]string{"label2"})
	g := LabelSet([]string{"label1", "label2"})
	h := LabelSet([]string{"label2", "label1"})
	i := LabelSet([]string{"label1", "label3"})
	tests := []tuple{
		tuple{a, a, true},
		tuple{a, b, true},
		tuple{b, a, true},
		tuple{c, d, true},
		tuple{c, b, false},
		tuple{a, e, false},
		tuple{g, h, true},
		tuple{h, i, false},
	}
	for _, test := range tests {
		actual := test.a.Equal(test.b)
		if actual != test.expect {
			t.Errorf("With sets %+v and %+v expected: %+v, actual %+v", test.a, test.b, test.expect, actual)
		}
	}
}
