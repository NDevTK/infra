package vector

import (
	"testing"
)

func TestVectorCompare(t *testing.T) {
	t.Parallel()

	a := Val(V{1, 0, 0})
	b := Val(V{1, 0, 1})
	c := Val(V{0, 1, 1})
	d := Val(V{1, 0, 0})
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
			t.Errorf("%+v < %+v ? expected: %+v actual: %+v",
				expect.A, expect.B, expect.Expect, actual)
		}
	}
}
