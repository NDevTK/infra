package types

// Mutater is an interface that represents mutations to State that the
// scheduler may emit.
type Mutater interface {
	Mutate(state *State)
}
