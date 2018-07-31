package types

import "sort"
// IsIdle returns whether the given worker is currently idle.
func (w *Worker) IsIdle() bool {
	return w.RunningTask == nil
}

// SortAscendingCost sorts a slice in-place by ascending cost.
func SortAscendingCost(ws []*Worker) {
	less := func(i, j int) bool {
		return ws[i].RunningTask.Cost.Less(*ws[j].RunningTask.Cost)
	}
	sort.SliceStable(ws, less)
}

// SortDescendingCost sorts a slice in-place by descending cost.
func SortDescendingCost(ws []*Worker) {
	less := func(i, j int) bool {
		return ws[j].RunningTask.Cost.Less(*ws[i].RunningTask.Cost)
	}
	sort.SliceStable(ws, less)
}
