// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"context"
	"sync"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"

	"infra/appengine/qscheduler-swarming/app/state/types"
	"infra/qscheduler/qslib/scheduler"
)

// BatchPriority is the priority level within a batch.
type BatchPriority int

const (
	// BatchPriorityNotify is the priority level of Notify requests.
	BatchPriorityNotify BatchPriority = iota

	// BatchPriorityAssign is the priority level of Assign requests.
	BatchPriorityAssign

	nBatchPriorities int = iota
)

// BatchRunner runs operations in batches.
//
// Requests within a batch are handled in a single read-modify-write
// transaction, in priority order.
//
// All public methods of BatchRunner are threadsafe.
type BatchRunner struct {
	// closed is closed to indicate that the Batcher has finished closing.
	closed chan struct{}

	// requests is the channel of operations to be run.
	requests chan *batchedOp

	// startOnce is used to ensure that the batcher is only started once.
	startOnce sync.Once

	// Test fixtures channels. These will always be initialized, but are
	// closed for non-test instances of Batcher, so that reads from them
	// succeed immediately without blocking.

	// tBatchWait is read from after a request is included in a batch.
	tBatchWait chan struct{}

	tWait bool

	// tBatchStart is read from prior to a batch being permitted to start.
	tBatchStart chan struct{}

	// doneChannelSize is the buffer size to use for done channels.
	//
	// In production, this is 1, to ensure that the single necessary write
	// to this channel doesn't block.
	//
	// In tests, this is 0, to ensure that batcher is deadlock-free.
	doneChannelSize int
}

// NewBatcher creates a new BatchRunner.
func NewBatcher() *BatchRunner {
	b := &BatchRunner{
		requests: make(chan *batchedOp, 100),
		closed:   make(chan struct{}),

		doneChannelSize: 1,

		tBatchStart: make(chan struct{}),
		tBatchWait:  make(chan struct{}),
	}
	b.closeFixtureChannels()
	return b
}

// Start starts a batcher (if it hasn't been started already).
//
// It returns immediately.
func (b *BatchRunner) Start(store *Store) {
	b.startOnce.Do(func() {
		go b.runRequestsInBatches(store)
	})
}

// RunOperation runs the given operation within a batch.
//
// Within a batch, operations are ordered by priority.
//
// RunOperation returns a channel that receives an error for the operation or
// closes once the operation has completed after which it is safe to read its
// result.
func (b *BatchRunner) RunOperation(ctx context.Context, op types.Operation, priority BatchPriority) (wait <-chan error) {
	// Use a buffered channel, so that writing back to this channel doesn't block.
	dc := make(chan error, b.doneChannelSize)
	bo := &batchedOp{
		ctx:       ctx,
		priority:  priority,
		operation: op,
		done:      dc,
	}

	// Attempt to join a batch, but bail out if context is cancelled.
	select {
	case b.requests <- bo:
		break
	case <-ctx.Done():
		dc <- ctx.Err()
		close(dc)
	}

	return dc
}

// Close closes a batcher, and waits for it to finish closing.
//
// Any requests that were previously enqueued to this batcher
// will be allowed to complete. Attempting to send any new requests
// to this batcher after calling Close will panic.
func (b *BatchRunner) Close() {
	b.closeFixtureChannels()
	close(b.requests)
	<-b.closed
}

// runRequestsInBatches creates new batches and runs them, until the requests
// channel closes.
func (b *BatchRunner) runRequestsInBatches(store *Store) {
	for r := range b.requests {
		<-b.tBatchStart

		// Create a new batch that will run in r's context.
		ctx := r.ctx
		logging.Debugf(ctx, "request picked as batch master")

		// If r's context is already dead, skip it and move on.
		select {
		case <-r.ctx.Done():
			// Request is already cancelled, don't use it as a master.
			logging.Debugf(ctx, "request already cancelled, dropped as batch master")
			continue
		default:
		}

		nb := &batch{}
		nb.append(r)
		<-b.tBatchWait

		b.collectForBatch(ctx, nb)
		logging.Debugf(ctx, "batch of size %s collected, executing", nb.numOperations())
		nb.executeAndClose(ctx, store)
		logging.Debugf(ctx, "batch executed")
	}
	// No more requests, close batches channel.
	close(b.closed)
}

func (b *BatchRunner) collectForBatch(ctx context.Context, nb *batch) {
	for {
		select {
		case r := <-b.requests:
			if r == nil {
				// Requests channel is closed, stop collecting.
				return
			}
			logging.Debugf(r.ctx, "request picked up as batch slave, will eventually execute")
			nb.append(r)
			<-b.tBatchWait
		case <-clock.After(ctx, 10*time.Millisecond):
			// Stop collecting, unless we are in a test test fixture and
			// waiting for additional requests.
			if !b.tWait {
				return
			}
		}
	}
}

// closeFixtureChannels closes all the channels related to the test fixture
// for Batcher. This causes Batcher to behave as though there were no test
// fixture.
func (b *BatchRunner) closeFixtureChannels() {
	close(b.tBatchStart)
	close(b.tBatchWait)
}

// batchedOp encapsulates a single operation to be batched.
type batchedOp struct {
	// ctx is the context of the originating request for this operation.
	//
	// It is examined and used only for the first operation of a batch, to be
	// used as the context that the entire batch runs in.
	//
	// Note: storing a context on a struct is discouraged by the golang docs;
	// in this case, the context is only being stored in order to be passed
	// through a channel and then be used as a parameter to batch.Build.
	ctx context.Context

	// operation is the Operation to be run.
	operation types.Operation

	// priority is the priority within the batch to run the operation.
	// Operations will be run within in the batch in ascending priority order.
	priority BatchPriority

	// err is the error that was encountered on the batch (so far) for this
	// operation.
	err error

	// done is a buffered channel, that should have the error for this operation written to it
	// or be closed if the operation completed without error.
	done chan<- error
}

// batch encapsulates a batch of operations.
type batch struct {
	// operations is (per-priority) collection of operations included in the batch.
	operations [nBatchPriorities][]*batchedOp
}

// append appends an operation to the batch.
func (b *batch) append(bo *batchedOp) {
	b.operations[bo.priority] = append(b.operations[bo.priority], bo)
}

func (b *batch) numOperations() int {
	count := 0
	for _, ops := range b.operations {
		count += len(ops)
	}
	return count
}

// executeAndClose executes and closes the given batch.
func (b *batch) executeAndClose(ctx context.Context, store *Store) {
	success := true
	if err := store.RunOperationInTransaction(ctx, b.getRunner(store)); err != nil {
		// A batch-wide error occurred. Store it on all results.
		b.allResultsError(err)
		success = false
	}
	recordBatchSize(ctx, b.numOperations(), store.entityID, success)

	b.close()
}

// getRunner gets a runner function to be used in a datastore transaction
// to execute the batch.
func (b *batch) getRunner(store *Store) types.Operation {
	return func(ctx context.Context, state *types.QScheduler, events scheduler.EventSink) error {
		// Modify
		for _, opSlice := range b.operations {
			for _, op := range opSlice {
				op.err = op.operation(ctx, state, events)
			}
		}
		return nil
	}
}

// allResultsError sets the given error to all operations in the batch.
func (b *batch) allResultsError(err error) {
	for _, opSlice := range b.operations {
		for _, op := range opSlice {
			op.err = err
		}
	}
}

// close closes a batch, sending out any necessary errors to operations.
func (b *batch) close() {
	for _, opSlice := range b.operations {
		for _, op := range opSlice {
			if op.err != nil {
				op.done <- op.err
			}

			close(op.done)
		}
	}
}

// TestFixtures

// NewBatchRunnerForTest creates a Batcher for testing purposes.
//
// On a batcher instance created with NewBatchRunnerForTest, the batcher requires calls
// to TBatchWait and TBatchClose to allow requests to be enqueued and for
// batches to be allowed to close.
func NewBatchRunnerForTest() *BatchRunner {
	return &BatchRunner{
		requests: make(chan *batchedOp),
		closed:   make(chan struct{}),

		doneChannelSize: 0,

		tBatchStart: make(chan struct{}),
		tBatchWait:  make(chan struct{}),
	}
}

// TBatchWait blocks until the given number of requests have been included in
// a batch.
//
// This is to be used only by tests, on Batcher instances created with
// NewForTest. Otherwise, this method panics.
func (b *BatchRunner) TBatchWait(requests int) {
	b.tWait = true
	for i := 0; i < requests; i++ {
		b.tBatchWait <- struct{}{}
	}
	b.tWait = false
}

// TBatchStart allows the currently building batch to stop building and start executing.
//
// This is to be used only by tests, on Batcher instances created with
// NewForTest. Otherwise, this method panics.
func (b *BatchRunner) TBatchStart() {
	b.tBatchStart <- struct{}{}
}
