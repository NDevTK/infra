/*
Package vector implements a protobuf-backed Vector of a compile-time known
length, used to store quota account values, as part of the quota scheduler
algorithm.
*/
package vector

// NumPriorities is the number of distinct priority buckets. For performance
// and code complexity reasons, this is a compile-time constant.
const NumPriorities = 3

// IntVector is the integer equivalent of QuotaVector, to store things
// like per-bucket counts.
type IntVector [NumPriorities]int

// EmptyVector creates an new 0-initialized Vector with the correct
// underlying slice size.
func EmptyVector() Vector{
	return Vector{Values: make([]float64, NumPriorities)}
}

// V is a convenient type alias for initializing Vectors with
// a value.
// e.g. Val(V{1, 2, 3}) for a Vector
type V [NumPriorities]float64

// Val is a convience method which creates a new Vector with
// initial values from val.
func Val(val V) Vector {
	return Vector{Values: val[:]}
}

// Ref is a convience method which creates a new *Vector with
// initial values from val.
func Ref(val V) *Vector {
	return &Vector{Values: val[:]}
}

// At is a convenience method to return a Vector's component at a given
// priority, without the caller needing to worry about bounds checks.
func (a Vector) At(priority int32) float64 {
	fix(&a)
	return a.Values[priority]
}

// fix ensures that a Vector's underlying slice has the correct length
// (NumPriorities) or fixes it accordingly.
func fix(v *Vector) {
	if len(v.Values) != NumPriorities {
		newSlice := make([]float64, NumPriorities)
		copy(newSlice, v.Values)
		v.Values = newSlice
	}
}

// Less determines whether Vector a is less than b, based on
// priority ordered comparison
func (a Vector) Less(b Vector) bool {
	fix(&a)
	fix(&b)
	for i, valA := range a.Values {
		valB := b.Values[i]
		if valA < valB {
			return true
		}
		if valB < valA {
			return false
		}
	}
	return false
}

// Plus returns the sum of two vectors.
func (a Vector) Plus(b Vector) Vector {
	ans := EmptyVector()
	copy(ans.Values, a.Values)
	for i, v := range b.Values {
		ans.Values[i] += v
	}
	return ans
}

// Minus returns the difference of two vectors.
func (a Vector) Minus(b Vector) Vector {
	ans := EmptyVector()
	copy(ans.Values, a.Values)
	for i, v := range b.Values {
		ans.Values[i] -= v
	}
	return ans
}

// Equals returns true if two given vectors are equal.
func (a Vector) Equals(b Vector) bool {
	fix(&a)
	fix(&b)
	for i, vA := range a.Values {
		if vA != b.Values[i] {
			return false
		}
	}
	return true
}
