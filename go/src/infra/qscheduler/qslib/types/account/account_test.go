package account

import (
	"testing"
)

const actualExpect = "expected: %+v, actual: %+v"

func TestBestPriorityForZeroBalance(t *testing.T) {
	t.Parallel()
	expect := FreeBucket
	actual := BestPriorityFor(&Balance{})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestBestPriorityForPositiveBalance(t *testing.T) {
	t.Parallel()
	expect := 1
	actual := BestPriorityFor(&Balance{0, 1, 0})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestAccountAdvanceWithNoOverflow(t *testing.T) {
	t.Parallel()
	expect := Balance{0, 2, 4}

	config := Config{
		ChargeRate: Vector{1, 2, 3},
		MaxBalance: Vector{10, 10, 10},
	}
	actual := Balance{}
	actual.Advance(&config, 2, IntVector{1, 1, 1})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}

func TestAccountAdvanceWithOverflow(t *testing.T) {
	t.Parallel()
	expect := Balance{10, 11, 10}
	// P0 bucket will start below max and reach max.
	// P1 bucket will have started above max already, but have spend that causes
	//    it to be pulled to a lower value still above max.
	// P2 bucket will have started above max, but have spend that causes it to be
	//    pulled below max, and then will recharge to reach max again.
	config := Config{
		ChargeRate: Vector{1, 1, 1},
		MaxBalance: Vector{10, 10, 10},
	}

	actual := Balance{9.5, 12, 10.5}
	actual.Advance(&config, 1, IntVector{0, 1, 1})

	if actual != expect {
		t.Errorf(actualExpect, expect, actual)
	}
}
