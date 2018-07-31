package types

import "infra/qscheduler/qslib/types/task"

// WorkerID is an opaque globally unique identifier of a particular worker
// (in Swarming, this corresponds conceptually to a Bot).
type WorkerID string

// Worker represents a resource that can run 1 task at a time. This corresponds
// to the swarming concept of a Bot. This representation considers only the
// subset of Labels that are Provisionable (can be changed by running a task),
// because the quota scheduler algorithm is expected to run against a pool of
// otherwise homogenous workers.
type Worker struct {
	ID          WorkerID
	Labels      task.LabelSet
	RunningTask *task.Running
}

// IsIdle returns whether the given worker is currently idle.
func (w *Worker) IsIdle() bool {
	return w.RunningTask == nil
}
