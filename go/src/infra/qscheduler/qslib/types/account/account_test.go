package account

import (
	"testing"
)

const actualExpect = "expected: %+v, actual: %+v"

func TestBestPriority(t *testing.T) {
	t.Parallel()
	expects := []int{
		FreeBucket,
		1,
	}
	actuals := []int{
		BestPriorityFor(&Vector{}),
		BestPriorityFor(&Vector{0, 1, 0}),
	}

	for i, expect := range expects {
		actual := actuals[i]
		if actual != expect {
			t.Errorf(actualExpect, expect, actual)
		}
	}
}

func TestBestPriorityForPositiveBalance(t *testing.T) {
	t.Parallel()
	expect := 1
	actual := BestPriorityFor(&Vector{0, 1, 0})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestAccountAdvanceWithNoOverflow(t *testing.T) {
	t.Parallel()
	expect := Vector{0, 2, 4}

	config := Config{
		ChargeRate: Vector{1, 2, 3},
		MaxBalance: Vector{10, 10, 10},
	}
	actual := Vector{}
	actual.Update(&config, 2, IntVector{1, 1, 1})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestAccountAdvanceWithOverflow(t *testing.T) {
	t.Parallel()
	expect := Vector{10, 11, 10}
	// P0 bucket will start below max and reach max.
	// P1 bucket will have started above max already, but have spend that causes
	//    it to be pulled to a lower value still above max.
	// P2 bucket will have started above max, but have spend that causes it to be
	//    pulled below max, and then will recharge to reach max again.
	config := Config{
		ChargeRate: Vector{1, 1, 1},
		MaxBalance: Vector{10, 10, 10},
	}

	actual := Vector{9.5, 12, 10.5}
	actual.Update(&config, 1, IntVector{0, 1, 1})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestVectorCompare(t *testing.T) {
	t.Parallel()
	a := Vector{1, 0, 0}
	b := Vector{1, 0, 1}
	c := Vector{0, 1, 1}
	d := Vector{1, 0, 0}
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
