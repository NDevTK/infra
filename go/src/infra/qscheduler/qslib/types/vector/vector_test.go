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
	type tuple struct {
		A      Vector
		B      Vector
		Expect bool
	}
	expects := []tuple{
		tuple{a, a, false},
		tuple{b, a, false},
		tuple{a, b, true},
		tuple{c, a, true},
		tuple{d, a, false},
	}
	for _, expect := range expects {
		if expect.A.Less(expect.B) != expect.Expect {
			t.Fail()
		}
	}
}
