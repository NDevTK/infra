// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package eval

import (
	"bytes"
	"container/heap"
	"context"
	"flag"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/rts"
	"infra/rts/presubmit/eval/history"
	evalpb "infra/rts/presubmit/eval/proto"
)

const defaultConcurrency = 100

// Eval estimates safety and efficiency of a given selection strategy.
type Eval struct {
	// The selection strategy to evaluate.
	Strategy Strategy

	// The number of goroutines to spawn for each metric.
	// If <=0, defaults to 100.
	Concurrency int

	// Rejections are used to evaluate safety of the strategy.
	Rejections *history.Player

	// Durations is a path to a directory with test duration records.
	// For format details, see comments of TestDurationRecord protobuf message.
	Durations string

	// LogFurthest instructs to log rejections for which failed tests have large
	// distance, as concluded by the selection strategy.
	// LogFurthest is the number of rejections to print, ordered by descending
	// distance.
	// This can help diagnosing the selection strategy.
	//
	// TODO(nodir): implement this.
	LogFurthest int
}

// RegisterFlags registers flags for the Eval fields.
func (e *Eval) RegisterFlags(fs *flag.FlagSet) error {
	fs.IntVar(&e.Concurrency, "j", defaultConcurrency, "Number of job to run parallel")
	fs.Var(&historyFileInputFlag{ptr: &e.Rejections}, "rejections", text.Doc(`
		Path to the history file with change rejections. Used for safety evaluation.
	`))
	fs.StringVar(&e.Durations, "durations", "", text.Doc(`
		Path to a directory with test duration records.
		For format details, see comments of TestDurationRecord protobuf message.
		Used for efficiency evaluation.
	`))
	fs.IntVar(&e.LogFurthest, "log-furthest", 10, text.Doc(`
		Log rejections for which failed tests have large distance,
		as concluded by the selection strategy.
		The flag value is the number of rejections to print, ordered by descending
		distance.
		This can help diagnosing the selection strategy.
	`))
	return nil
}

// ValidateFlags validates values of flags registered using RegisterFlags.
func (e *Eval) ValidateFlags() error {
	if e.Rejections == nil {
		return errors.New("-rejections is required")
	}
	if e.Durations == "" {
		return errors.New("-durations is required")
	}
	return nil
}

// Run evaluates the candidate strategy.
func (e *Eval) Run(ctx context.Context) (*Result, error) {
	res := &Result{}

	logging.Infof(ctx, "evaluating safety...")
	grid, err := e.evaluateSafety(ctx, res)
	if err != nil {
		return nil, errors.Annotate(err, "failed to evaluate safety").Err()
	}

	logging.Infof(ctx, "evaluating efficiency...")
	if err := e.evaluateEfficiency(ctx, res, grid); err != nil {
		return nil, errors.Annotate(err, "failed to evaluate efficiency").Err()
	}

	res.Thresholds = grid.Slice()
	return res, nil
}

// evaluateSafety computes thresholds and total/preserved
// rejections/testFailures.
func (e *Eval) evaluateSafety(ctx context.Context, res *Result) (*thresholdGrid, error) {
	var changeAffectedness []rts.Affectedness
	var testAffectedness []rts.Affectedness
	furthest := make(furthestRejections, 0, e.LogFurthest)
	var mu sync.Mutex

	eg, ctx := errgroup.WithContext(ctx)
	defer eg.Wait()

	// Play back the history.
	eg.Go(func() error {
		err := e.Rejections.PlaybackRejections(ctx)
		return errors.Annotate(err, "failed to playback history").Err()
	})

	e.goMany(eg, func() error {
		for rej := range e.Rejections.RejectionC {
			// TODO(crbug.com/1112125): skip the patchset if it has a ton of failed tests.
			// Most selection strategies would reject such a patchset, so it represents noise.

			// Invoke the strategy.
			in := Input{TestVariants: rej.FailedTestVariants}
			in.ensureChangedFilesInclude(rej.Patchsets...)
			out := &Output{TestVariantAffectedness: make([]rts.Affectedness, len(in.TestVariants))}
			if err := e.Strategy(ctx, in, out); err != nil {
				return errors.Annotate(err, "the selection strategy failed").Err()
			}

			// The affectedness of a change is based on the most affected failed test.
			mostAffected, err := mostAffected(out.TestVariantAffectedness)
			if err != nil {
				return err
			}

			mu.Lock()
			changeAffectedness = append(changeAffectedness, mostAffected)
			testAffectedness = append(testAffectedness, out.TestVariantAffectedness...)
			furthest.Consider(affectedRejection{Rejection: rej, MostAffected: mostAffected})
			mu.Unlock()
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if len(changeAffectedness) == 0 {
		return nil, errors.New("no change rejections")
	}

	// Initialize the grid.
	// Compute distance/rank thresholds by taking distance/rank percentiles in
	// changeAffectedness. Row/column indexes represent ChangeRecall scores.
	grid := &thresholdGrid{}
	distancePercentiles, rankPercentiles := grid.init(changeAffectedness)
	logging.Infof(ctx, "Distance percentiles: %v", distancePercentiles)
	logging.Infof(ctx, "Rank percentiles: %v", rankPercentiles)
	if len(furthest) > 0 {
		furthest.LogAndClear(ctx)
	}

	// Evaluate safety of *combinations* of thresholds.

	losses := func(afs []rts.Affectedness) *bucketGrid {
		var buckets bucketGrid
		for _, af := range afs {
			buckets.inc(grid, af, 1)
		}
		buckets.makeCumulative()
		return &buckets
	}

	res.TotalRejections = len(changeAffectedness)
	lostRejections := losses(changeAffectedness)

	res.TotalTestFailures = len(testAffectedness)
	lostFailures := losses(testAffectedness)

	for row := 0; row < 100; row++ {
		for col := 0; col < 100; col++ {
			t := &grid[row][col]
			t.PreservedRejections = res.TotalRejections - int(lostRejections[row+1][col+1])
			t.PreservedTestFailures = res.TotalTestFailures - int(lostFailures[row+1][col+1])
			t.ChangeRecall = float64(t.PreservedRejections) / float64(res.TotalRejections)
			t.DistanceChangeRecall = float64(row+1) / 100
			t.RankChangeRecall = float64(col+1) / 100
			t.TestRecall = float64(t.PreservedTestFailures) / float64(res.TotalTestFailures)
		}
	}
	return grid, nil
}

// evaluateEfficiency computes total and saved durations.
func (e *Eval) evaluateEfficiency(ctx context.Context, res *Result, grid *thresholdGrid) error {
	// Process test durations in parallel and increment appropriate counters.
	var savedDurations bucketGrid
	var totalDuration int64

	eg, ctx := errgroup.WithContext(ctx)
	defer eg.Wait()

	// Play back the history.
	recordC := make(chan *evalpb.TestDurationRecord)
	eg.Go(func() error {
		defer close(recordC)
		err := readTestDurations(ctx, e.Durations, recordC)
		return errors.Annotate(err, "failed to read test duration records").Err()
	})

	e.goMany(eg, func() error {
		in := Input{}
		out := &Output{}
		for rec := range recordC {
			// Invoke the strategy.
			if cap(in.TestVariants) < len(rec.TestDurations) {
				in.TestVariants = make([]*evalpb.TestVariant, len(rec.TestDurations))
			}
			in.TestVariants = in.TestVariants[:len(rec.TestDurations)]
			for i, td := range rec.TestDurations {
				in.TestVariants[i] = td.TestVariant
			}
			in.ChangedFiles = in.ChangedFiles[:0]
			in.ensureChangedFilesInclude(rec.Patchsets...)

			out.TestVariantAffectedness = make([]rts.Affectedness, len(in.TestVariants))
			if err := e.Strategy(ctx, in, out); err != nil {
				return errors.Annotate(err, "the selection strategy failed").Err()
			}

			// Record results.
			durSum := int64(0)
			for i, td := range rec.TestDurations {
				dur := int64(td.Duration.AsDuration())
				durSum += dur
				savedDurations.inc(grid, out.TestVariantAffectedness[i], dur)
			}
			atomic.AddInt64(&totalDuration, durSum)
		}
		return ctx.Err()
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	if totalDuration == 0 {
		return errors.New("sum of test durations is 0")
	}

	// Incroporate the counters into res.

	res.TotalDuration = time.Duration(totalDuration)
	savedDurations.makeCumulative()
	for row := 0; row < 100; row++ {
		for col := 0; col < 100; col++ {
			t := &grid[row][col]
			t.SavedDuration = time.Duration(savedDurations[row+1][col+1])
			t.Savings = float64(t.SavedDuration) / float64(res.TotalDuration)
		}
	}
	return nil
}

func (e *Eval) goMany(eg *errgroup.Group, f func() error) {
	concurrency := e.Concurrency
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	for i := 0; i < concurrency; i++ {
		eg.Go(func() error {
			return f()
		})
	}
}

// thresholdGrid is a 100x100 grid where each cell represents a distance/rank
// threshold. All cells in the same row have the same distance value, and
// all cells in the same column have the same rank value.
//
// The distance value of the row R is the minimal distance threshold required to
// achieve ChangeRecall score of (R+1)/100.0 on the training set, while ignoring
// rank threshold.
// For example, thresholdGrid[94][0].Value.Distance achieves 95% ChangeRecall
// on the training set.
//
// Similarly, rank value of the column C is the minimal rank threshold required
// to achieve ChangeRecall score of (C+1)/100.0 on to the training set,
// while ignoring distance threshold.
//
// The distances and ranks are computed independently of each other and then
// combined into this grid.
type thresholdGrid [100][100]Threshold

// init clears the grid and initializes the rows/columns with distance/rank
// percentiles in afs.
func (g *thresholdGrid) init(afs []rts.Affectedness) (distancePercentiles []float64, rankPercentiles []int) {
	*g = thresholdGrid{}
	distancePercentiles, rankPercentiles = affectednessQuantiles(afs, 100)
	for row, distance := range distancePercentiles {
		for col, rank := range rankPercentiles {
			g[row][col].Value = rts.Affectedness{Distance: distance, Rank: rank}
		}
	}
	return
}

// Slice returns a slice of thresholds, sorted by ChangeRecall.
// If two thresholds have the same ChangeScore, the one with larger Savings
// wins. Returned slice elements are pointes to the grid.
func (g *thresholdGrid) Slice() []*Threshold {
	// Given that the grid has 100 cells for each distance threshold and
	// each rank threshold, many cells have the same ChangeRecall.
	// Deduplicate them while breaking the tie based on Savings.
	best := map[float64]*Threshold{}
	for row := 0; row < 100; row++ {
		for col := 0; col < 100; col++ {
			t := &g[row][col]
			if existing := best[t.ChangeRecall]; existing == nil || t.Savings > existing.Savings {
				best[t.ChangeRecall] = t
			}
		}
	}

	keys := make([]float64, 0, len(best))
	for score := range best {
		keys = append(keys, score)
	}
	sort.Float64s(keys)

	ret := make([]*Threshold, len(best))
	for i, key := range keys {
		ret[i] = best[key]
	}
	return ret
}

// quantiles returns distance and rank quantiles.
// Panics if s is empty.
func affectednessQuantiles(afs []rts.Affectedness, count int) (distances []float64, ranks []int) {
	if len(afs) == 0 {
		panic("s is empty")
	}
	allDistances := make([]float64, len(afs))
	allRanks := make([]int, len(afs))
	for i, af := range afs {
		allDistances[i] = af.Distance
		allRanks[i] = af.Rank
	}
	sort.Float64s(allDistances)
	sort.Ints(allRanks)
	distances = make([]float64, count)
	ranks = make([]int, count)
	for i := 0; i < count; i++ {
		boundary := int(math.Ceil(float64(len(afs)*(i+1)) / float64(count)))
		distances[i] = allDistances[boundary-1]
		ranks[i] = allRanks[boundary-1]
	}
	return
}

// bucketGrid is an auxulary data structure to compute cumulative counters
// in ThresholdGrid. Each cell contains the number of data points lost by that
// bucket.
//
// bucketGrid is used in two phases:
//   1) For each data point, call inc().
//   2) Call makeCumulative() and incorporate bucketGrid into ThresholdGrid.
//
// The structure of bucketGrid is similar to ThresholdGrid, except bucketGrid
// cell (R, C) corresponds to ThresholdGrid cell (R-1, C-1). This is because the
// bucketGrid is padded with extra row 0 and column 0 for data points that were
// not lost by any distance or by any rank.
type bucketGrid [101][101]int64

// inc increments the counter in the cell (R, C) where row R has the largest
// distance less than af.Distance, and C has the largest rank less than af.Rank.
// In other words, it increments the largest thresholds that missed the data
// point.
//
// Goroutine-safe.
func (b *bucketGrid) inc(g *thresholdGrid, af rts.Affectedness, delta int64) {
	row := sort.Search(100, func(i int) bool {
		return g[i][0].Value.Distance >= af.Distance
	})
	col := sort.Search(100, func(i int) bool {
		return g[0][i].Value.Rank >= af.Rank
	})

	// For each of distance and rank, we have found the index of the smallest
	// threshold satisfied by af.
	// We need the *largest* threshold *not* satisfied by af, i.e. the preceding
	// index. Indexes in bucketGrid are already shifted by one, so use row and
	// col as is.
	atomic.AddInt64(&b[row][col], delta)
}

// makeCumulative makes all counters cumulative.
// Not idempotent.
func (b *bucketGrid) makeCumulative() {
	for row := 99; row >= 0; row-- {
		for col := 99; col >= 0; col-- {
			b[row][col] += b[row][col+1]
		}
	}
	for col := 99; col >= 0; col-- {
		for row := 99; row >= 0; row-- {
			b[row][col] += b[row+1][col]
		}
	}
}

// mostAffected returns the most significant Affectedness, in terms of distance
// and rank.
// If two tests have the same distance and ranks are available, then ranks are
// used to break the tie.
//
// Returns an error if an inconsistency is found among ranks and distances,
// see also checkConsistency.
func mostAffected(afs []rts.Affectedness) (rts.Affectedness, error) {
	if len(afs) == 0 {
		return rts.Affectedness{}, errors.New("empty")
	}
	most := afs[0]
	for _, af := range afs {
		if err := checkConsistency(most, af); err != nil {
			return rts.Affectedness{}, err
		}
		haveRanks := af.Rank != 0 && most.Rank != 0
		if most.Distance > af.Distance || haveRanks && most.Rank > af.Rank {
			most = af
		}
	}
	return most, nil
}

// checkConsistency returns a non-nil error if the order implied by ranks
// and distances is inconsistent in a and b, e.g. a is closer, but its rank
// is greater.
func checkConsistency(a, b rts.Affectedness) error {
	if a.Rank == 0 || b.Rank == 0 {
		return nil
	}
	if (a.Distance < b.Distance && a.Rank >= b.Rank) || (a.Distance > b.Distance && a.Rank <= b.Rank) {
		return errors.Reason("ranks and distances are inconsistent: %#v and %#v", a, b).Err()
	}
	return nil
}

type furthestRejections []affectedRejection
type affectedRejection struct {
	Rejection    *evalpb.Rejection
	MostAffected rts.Affectedness
}

// Consider pushes the item if there is unused capacity, or replaces the closest
// item in the heap if the former is further than the latter.
// Does nothing if cap(*f) == 0.
func (f *furthestRejections) Consider(item affectedRejection) {
	switch {
	case cap(*f) == 0:
		return

	// If the heap has a free slot, just add the rejection.
	case len(*f) < cap(*f):
		heap.Push(f, item)

	// Otherwise, if the rejection is further than heap's closest one, then
	// replace the latter.
	case (*f)[0].MostAffected.Distance < item.MostAffected.Distance:
		(*f)[0] = item
		heap.Fix(f, 0)
	}
}

func (f *furthestRejections) LogAndClear(ctx context.Context) {
	buf := &bytes.Buffer{}
	p := rejectionPrinter{printer: newPrinter(buf)}
	p.printf("%d furthest rejections:\n", len(*f))
	p.Level++
	for len(*f) > 0 {
		r := heap.Pop(f).(affectedRejection)
		p.rejection(r.Rejection, r.MostAffected)
	}
	p.Level--
	logging.Infof(ctx, "%s", buf.Bytes())
}

func (f furthestRejections) Len() int { return len(f) }
func (f furthestRejections) Less(i, j int) bool {
	return f[i].MostAffected.Distance < f[j].MostAffected.Distance
}
func (f furthestRejections) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f *furthestRejections) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*f = append(*f, x.(affectedRejection))
}
func (f *furthestRejections) Pop() interface{} {
	old := *f
	n := len(old)
	x := old[n-1]
	*f = old[0 : n-1]
	return x
}
