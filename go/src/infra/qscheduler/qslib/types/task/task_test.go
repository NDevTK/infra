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
		actual := test.a.Equals(test.b)
		if actual != test.expect {
			t.Errorf("With sets %+v and %+v expected: %+v, actual %+v", test.a, test.b, test.expect, actual)
		}
	}
}
