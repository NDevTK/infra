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

	"github.com/golang/protobuf/ptypes"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"

	"infra/appengine/qscheduler-swarming/app/config"
	"infra/appengine/qscheduler-swarming/app/state/metrics"
	"infra/appengine/qscheduler-swarming/app/state/nodestore"
	"infra/appengine/qscheduler-swarming/app/state/operations"
	"infra/appengine/qscheduler-swarming/app/state/types"
	"infra/qscheduler/qslib/scheduler"
	swarming "infra/swarming"
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
	requests chan batchable

	// startOnce is used to ensure that the batcher is only started once.
	startOnce sync.Once

	// Test fixtures channels.

	// tBatchWait is read from after a request is included in a batch.
	// This is closed in non-test instance of Batcher, so that reads always
	// succeed immediately without blocking.
	tBatchWait chan struct{}

	// A write to tBatchStart causes a batch to stop constructing and start
	// executing.
	tBatchStart chan struct{}

	// doneChannelSize is the buffer size to use for done channels.
	//
	// In production, this is 1, to ensure that the single necessary write
	// to this channel doesn't block.
	//
	// In tests, this is 0, to ensure that batcher is deadlock-free.
	doneChannelSize int

	poolID string
}

// NewBatcher creates a new BatchRunner.
func NewBatcher(poolID string) *BatchRunner {
	b := &BatchRunner{
		poolID: poolID,

		requests: make(chan batchable, 100),
		closed:   make(chan struct{}),

		doneChannelSize: 1,

		tBatchWait: make(chan struct{}),
	}
	b.closeFixtureChannels()
	return b
}

// Start starts a batcher (if it hasn't been started already).
//
// It returns immediately.
func (b *BatchRunner) Start(store *nodestore.NodeStore) {
	b.startOnce.Do(func() {
		go b.runRequestsInBatches(store)
	})
}

// Notify runs the given notify request in a batch.
func (b *BatchRunner) Notify(ctx context.Context, req *swarming.NotifyTasksRequest) (*swarming.NotifyTasksResponse, error) {
	bn := b.enqueueNotify(ctx, req)
	err, ok := <-bn.done
	if ok {
		return nil, err
	}
	return bn.resp, nil
}

// Assign runs the given assign request in a batch.
func (b *BatchRunner) Assign(ctx context.Context, req *swarming.AssignTasksRequest) (*swarming.AssignTasksResponse, error) {
	ba := b.enqueueAssign(ctx, req)
	err, ok := <-ba.done
	if ok {
		return nil, err
	}
	return ba.resp, nil
}

func (b *BatchRunner) enqueueNotify(ctx context.Context, req *swarming.NotifyTasksRequest) *batchedNotify {
	// Use a buffered channel, so that writing back to this channel doesn't block.
	dc := make(chan error, b.doneChannelSize)
	bo := &batchedNotify{
		batchedOp: batchedOp{
			ctx:  ctx,
			done: dc,
		},
		req: req,
	}

	go b.tryJoin(ctx, bo)

	return bo
}

func (b *BatchRunner) enqueueAssign(ctx context.Context, req *swarming.AssignTasksRequest) *batchedAssign {
	// Use a buffered channel, so that writing back to this channel doesn't block.
	dc := make(chan error, b.doneChannelSize)
	bo := &batchedAssign{
		batchedOp: batchedOp{
			ctx:  ctx,
			done: dc,
		},
		req: req,
	}

	go b.tryJoin(ctx, bo)

	return bo
}

func (b *BatchRunner) tryJoin(ctx context.Context, bo batchable) {
	select {
	case <-ctx.Done():
		bo.SetError(ctx.Err())
		bo.Close()
	case b.requests <- bo:
	}
}

func isValidOp(op interface{}) bool {
	switch op.(type) {
	case *swarming.AssignTasksRequest:
		return true
	case *swarming.NotifyTasksRequest:
		return true
	case types.Operation:
		return true
	default:
		return false
	}
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
func (b *BatchRunner) runRequestsInBatches(store *nodestore.NodeStore) {
	for r := range b.requests {

		// Create a new batch that will run in r's context.
		ctx := r.Ctx()
		logging.Debugf(ctx, "request picked as batch master")

		if !r.isActive() {
			// Request is already cancelled, don't use it as a master.
			logging.Debugf(ctx, "request already cancelled, dropped as batch master")
			continue
		}

		nb := &batch{}
		nb.append(r)
		// In test fixture, wait for a signal to continue after appending
		// an item to the batch.
		// In production, this channel is closed to the read returns immediately.
		<-b.tBatchWait

		err := b.collectForBatch(ctx, nb)
		if err != nil {
			logging.Debugf(ctx, "batch of size %d cancelled due to error", nb.numOperations(), err)
			nb.allResultsError(err)
			nb.close()
			continue
		}

		logging.Debugf(ctx, "batch of size %d collected, executing", nb.numOperations())
		nb.executeAndClose(ctx, store, b.poolID)
		logging.Debugf(ctx, "batch executed")
	}
	// No more requests, close batches channel.
	close(b.closed)
}

func (b *BatchRunner) collectForBatch(ctx context.Context, nb *batch) error {
	timer := clock.After(ctx, waitToCollect(ctx))
	for {
		select {
		case r := <-b.requests:
			if r == nil {
				// Requests channel is closed, stop collecting.
				return nil
			}
			if !r.isActive() {
				logging.Debugf(r.Ctx(), "request already cancelled, ignored for batch")
				continue
			}
			logging.Debugf(r.Ctx(), "request picked up as batch slave, will eventually execute")
			nb.append(r)
			// In test fixture, wait for a signal to continue after appending
			// an item to the batch.
			// In production, this channel is closed to the read returns immediately.
			<-b.tBatchWait
		case <-ctx.Done():
			return ctx.Err()
		case tr := <-timer:
			return tr.Err
		case <-b.tBatchStart:
			// Stop collecting due to test fixture signal.
			// In production, this codepath is never followed.
			return nil
		}
	}
}

const defaultConstructionWait = 300 * time.Millisecond

func waitToCollect(ctx context.Context) time.Duration {
	c := config.Get(ctx)
	if c == nil || c.QuotaScheduler == nil || c.QuotaScheduler.BatchConstructionWait == nil {
		return defaultConstructionWait
	}
	wait, err := ptypes.Duration(c.QuotaScheduler.BatchConstructionWait)
	if err != nil {
		return defaultConstructionWait
	}
	return wait
}

// closeFixtureChannels closes all the channels related to the test fixture
// for Batcher. This causes Batcher to behave as though there were no test
// fixture.
func (b *BatchRunner) closeFixtureChannels() {
	close(b.tBatchWait)
}

type batchable interface {
	Close()
	SetError(error)
	Ctx() context.Context
	isActive() bool
}

// batchedOp represents single operation to be batched.
//
// batchedOp methods and fields are not concurrency-safe.
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

	// err is the error that was encountered on the batch (so far) for this
	// operation.
	err error

	// done is a buffered channel, that will have the error for this operation
	// written to it or be closed if the operation completed without error.
	done chan error
}

func (b *batchedOp) Close() {
	if b.err != nil {
		b.done <- b.err
	}
	close(b.done)
}

func (b *batchedOp) Ctx() context.Context {
	return b.ctx
}

func (b *batchedOp) SetError(err error) {
	b.err = err
}

// isActive returns true if this operation is still active (its context is not
// cancelled).
func (b *batchedOp) isActive() bool {
	select {
	case <-b.ctx.Done():
		return false
	default:
		return true
	}
}

type batchedNotify struct {
	batchedOp

	req  *swarming.NotifyTasksRequest
	resp *swarming.NotifyTasksResponse
}

type batchedAssign struct {
	batchedOp

	req  *swarming.AssignTasksRequest
	resp *swarming.AssignTasksResponse
}

// batch encapsulates a batch of operations.
type batch struct {
	// notifyRequests is the set of NotifyRequest operations included in this
	// batch.
	notifyRequests []*batchedNotify
	// assignRequests is the set of AssignRequest operations included in this
	// batch.
	assignRequests []*batchedAssign

	count int
}

// append appends an operation to the batch.
func (b *batch) append(bo batchable) {
	switch o := bo.(type) {
	case *batchedNotify:
		b.notifyRequests = append(b.notifyRequests, o)
	case *batchedAssign:
		b.assignRequests = append(b.assignRequests, o)
	default:
		panic("invalid operation type appended to batch")
	}

	b.count++
}

func (b *batch) numOperations() int {
	return b.count
}

// executeAndClose executes and closes the given batch.
func (b *batch) executeAndClose(ctx context.Context, store *nodestore.NodeStore, poolID string) {
	success := true
	nodeRunner := NewNodeStoreOperationRunner(b.getRunner(), poolID)

	if err := store.Run(ctx, nodeRunner); err != nil {
		// A batch-wide error occurred. Store it on all results.
		b.allResultsError(err)
		success = false
	}
	metrics.RecordBatchSize(ctx, b.numOperations(), poolID, success)

	b.close()
}

func (b *batch) allOps() []batchable {
	all := make([]batchable, len(b.notifyRequests)+len(b.assignRequests))
	i := 0
	for _, n := range b.notifyRequests {
		all[i] = n
		i++
	}
	for _, a := range b.assignRequests {
		all[i] = a
		i++
	}
	return all
}

// getRunner gets a types.Operation that runs the batch.
func (b *batch) getRunner() types.Operation {
	return func(ctx context.Context, state *types.QScheduler, events scheduler.EventSink) {
		// Run all notify requests in individual operations; there is no
		// overhead improving in combining them to a single operation.
		for _, notify := range b.notifyRequests {
			op, resp := operations.NotifyTasks(notify.req)
			op(ctx, state, events)
			notify.resp = resp
		}

		// Run all assign requests in a single operation, so that they all
		// run in a single pass of the scheduler algorithm.
		assignReqs := make([]*swarming.AssignTasksRequest, len(b.assignRequests))
		for i, assign := range b.assignRequests {
			assignReqs[i] = assign.req
		}

		op, resp := operations.AssignTasks(assignReqs)
		op(ctx, state, events)

		for i, assign := range b.assignRequests {
			assign.resp = resp[i]
		}
	}
}

// allResultsError sets the given error to all operations in the batch.
func (b *batch) allResultsError(err error) {
	all := b.allOps()
	for _, op := range all {
		op.SetError(err)
	}
}

// close closes a batch, sending out any necessary errors to operations.
func (b *batch) close() {
	all := b.allOps()
	for _, op := range all {
		op.Close()
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
		requests: make(chan batchable),
		closed:   make(chan struct{}),

		doneChannelSize: 0,

		tBatchWait:  make(chan struct{}),
		tBatchStart: make(chan struct{}),
	}
}

// TBatchWait blocks until the given number of requests have been included in
// a batch.
//
// This is to be used only by tests, on Batcher instances created with
// NewForTest. Otherwise, this method panics.
func (b *BatchRunner) TBatchWait(requests int) {
	for i := 0; i < requests; i++ {
		b.tBatchWait <- struct{}{}
	}
}

// TBatchStart allows a new batch to start executing, and blocks until it does
// so.
//
// This is to be used only by tests, on Batcher instances created with
// NewForTest. Otherwise, this method panics.
func (b *BatchRunner) TBatchStart() {
	b.tBatchStart <- struct{}{}
}
