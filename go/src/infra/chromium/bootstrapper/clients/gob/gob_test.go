// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gob

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeClient struct {
	err   error
	max   int
	count int
}

func (c *fakeClient) op() error {
	c.count += 1
	if c.max < 0 || c.count <= c.max {
		return c.err
	}
	return nil
}

func messageForCode(code codes.Code) string {
	return fmt.Sprintf("fake %s error", code)
}

func TestExecute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = UseTestClock(ctx)

	Convey("Execute", t, func() {

		var retriableErrors = []codes.Code{
			codes.NotFound,
			codes.Unavailable,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
		}

		Convey("does not retry on errors without status code", func() {
			client := &fakeClient{
				err: errors.New("fake error without code"),
				max: 1,
			}

			err := Execute(ctx, "fake op", client.op)

			So(err, ShouldErrLike, "fake error without code")
		})

		Convey("does not retry on errors with non-retriable code", func() {
			client := &fakeClient{
				err: status.Error(codes.InvalidArgument, "fake error with non-retriable code"),
				max: 1,
			}

			err := Execute(ctx, "fake op", client.op)

			So(err, ShouldErrLike, "fake error with non-retriable code")
		})

		for _, code := range retriableErrors {
			Convey(fmt.Sprintf("retries on %s errors", code), func() {
				message := messageForCode(code)
				client := &fakeClient{
					err: status.Error(code, message),
					max: 1,
				}

				err := Execute(ctx, "fake op", client.op)

				So(err, ShouldBeNil)

				Convey("unless retries are disabled", func() {
					client.count = 0
					ctx := DisableRetries(ctx)

					err := Execute(ctx, "fake op", client.op)

					So(err, ShouldErrLike, message)
				})

				Convey("when retries are re-enabled", func() {
					client.count = 0
					ctx := EnableRetries(DisableRetries(ctx))

					err := Execute(ctx, "fake op", client.op)

					So(err, ShouldBeNil)
				})
			})
		}

		Convey("fails if operation does not succeed within max time", func() {
			code := retriableErrors[0]
			message := messageForCode(code)
			client := &fakeClient{
				err: status.Error(code, message),
				max: -1,
			}

			err := Execute(ctx, "fake op", client.op)

			So(err, ShouldErrLike, message)
		})

	})
}
