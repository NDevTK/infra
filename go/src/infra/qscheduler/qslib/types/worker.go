package types

// IsIdle returns whether the given worker is currently idle.
func (w *Worker) IsIdle() bool {
	return w.RunningTask == nil
}
