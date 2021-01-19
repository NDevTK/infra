// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package history

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/common/errors"

	evalpb "infra/rts/presubmit/eval/proto"
)

// Player can playback a history from a reader.
// It takes care of reconstructing rejections from fragments.
type Player struct {
	RejectionC chan *evalpb.Rejection
	DurationC  chan *evalpb.TestDuration
	r          *Reader
}

// NewPlayer returns a new Player.
func NewPlayer(r *Reader) *Player {
	return &Player{
		r:          r,
		RejectionC: make(chan *evalpb.Rejection),
		DurationC:  make(chan *evalpb.TestDuration),
	}
}

// Playback reads historical records and dispatches them to p.RejectionC and
// p.DurationC.
// Before exiting, closes RejectionC, DurationC and the underlying reader.
//
// TODO(nodir): refactor this package and potentially the file format.
// This function is never used.
func (p *Player) Playback(ctx context.Context) error {
	return p.playback(ctx, true, true)
}

// PlaybackRejections is like Playback, except reports only rejections.
func (p *Player) PlaybackRejections(ctx context.Context) error {
	return p.playback(ctx, false, true)
}

// PlaybackDurations is like Playback, except reports only durations.
func (p *Player) PlaybackDurations(ctx context.Context) error {
	return p.playback(ctx, true, false)
}

func (p *Player) playback(ctx context.Context, withDurations, withRejections bool) error {
	defer func() {
		close(p.RejectionC)
		close(p.DurationC)
		p.r.Close()
	}()

	curRej := &evalpb.Rejection{}
	for {
		rec, err := p.r.Read()
		switch {
		case err == io.EOF:
			return p.r.Close()
		case err != nil:
			return errors.Annotate(err, "failed to read history").Err()
		}

		// Send the record to the appropriate channel.
		switch data := rec.Data.(type) {

		case *evalpb.Record_RejectionFragment:
			if withRejections {
				proto.Merge(curRej, data.RejectionFragment.Rejection)
				if data.RejectionFragment.Terminal {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case p.RejectionC <- curRej:
						// Start a new rejection.
						curRej = &evalpb.Rejection{}
					}
				}
			}

		case *evalpb.Record_TestDuration:
			if withDurations {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case p.DurationC <- data.TestDuration:
				}
			}

		default:
			panic(fmt.Sprintf("unexpected record %s", rec))
		}
	}
}
