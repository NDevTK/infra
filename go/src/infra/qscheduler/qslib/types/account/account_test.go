package account

import (
	"testing"
	. "infra/qscheduler/qslib/types/vector"
)

const actualExpect = "expected: %+v, actual: %+v"

func TestBestPriority(t *testing.T) {
	t.Parallel()
	expects := []int32{
		FreeBucket,
		1,
	}
	actuals := []int32{
		BestPriorityFor(EmptyVector()),
		BestPriorityFor(Val(V{0, 1, 0})),
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
	expect := int32(1)
	actual := BestPriorityFor(Val(V{0, 1, 0}))

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestAccountAdvanceWithNoOverflow(t *testing.T) {
	t.Parallel()
	expect := Val(V{0, 2, 4})

	config := Config{
		ChargeRate: Ref(V{1, 2, 3}),
		MaxBalance: Ref(V{10, 10, 10}),
	}
	actual := EmptyVector()
	UpdateBalance(&actual, config, 2, &IntVector{1, 1, 1})

	if !actual.Equals(expect) {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestAccountAdvanceWithOverflow(t *testing.T) {
	t.Parallel()
	expect := Val(V{10, 11, 10})
	// P0 bucket will start below max and reach max.
	// P1 bucket will have started above max already, but have spend that causes
	//    it to be pulled to a lower value still above max.
	// P2 bucket will have started above max, but have spend that causes it to be
	//    pulled below max, and then will recharge to reach max again.
	config := Config{
		ChargeRate: Ref(V{1, 1, 1}),
		MaxBalance: Ref(V{10, 10, 10}),
	}

	actual := Val(V{9.5, 12, 10.5})
	UpdateBalance(&actual, config, 1, &IntVector{0, 1, 1})

	if !actual.Equals(expect) {
		t.Errorf(actualExpect, expect, actual)
	}
}

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
