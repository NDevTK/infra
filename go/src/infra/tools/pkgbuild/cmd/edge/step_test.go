// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// We only test the builder on a subset of platforms we support.
// Other platforms should be cross-compiled.
//go:build amd64 || (arm64 && darwin)

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/luciexe/build"
)

func TestRootStep(t *testing.T) {
	Convey("root step", t, func() {
		ctx := context.Background()
		s := NewRootStep(ctx, "root", "id")

		Convey("ok", func() {
			So(s.ID(), ShouldEqual, "id")

			executed := false
			err := s.RunSubstep(ctx, func(ctx context.Context, root *build.Step) error {
				executed = true
				return nil
			})
			So(err, ShouldBeNil)
			So(executed, ShouldBeTrue)

			s.End()
			s.End() // idempotent
			So(s.IsEnded(), ShouldBeTrue)
		})

		Convey("err", func() {
			err := s.RunSubstep(ctx, func(ctx context.Context, root *build.Step) error {
				return fmt.Errorf("failed1")
			})
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "failed1")

			err = s.RunSubstep(ctx, func(ctx context.Context, root *build.Step) error {
				return fmt.Errorf("failed2")
			})
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "failed2")

			err = s.RunSubstep(ctx, func(ctx context.Context, root *build.Step) error {
				return nil
			})
			So(err, ShouldBeNil)

			So(fmt.Sprintf("%s", s.Err()), ShouldContainSubstring, "failed1")
			So(fmt.Sprintf("%s", s.Err()), ShouldContainSubstring, "failed2")

			s.End()

			err = s.RunSubstep(ctx, func(ctx context.Context, root *build.Step) error {
				return nil
			})
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "ended")

			So(s.IsEnded(), ShouldBeTrue)
		})

		Convey("cancel", func() {
			So(s.ID(), ShouldEqual, "id")

			ctx, cancel := context.WithCancel(ctx)
			done := make(chan struct{})

			var err error
			go func() {
				err = s.RunSubstep(ctx, func(ctx context.Context, root *build.Step) error {
					<-ctx.Done()
					return nil
				})
				close(done)
			}()
			cancel()
			<-done
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "cancled")
		})
	})
}

func TestRootSteps(t *testing.T) {
	Convey("root steps", t, func() {
		ctx := context.Background()
		s := NewRootSteps()

		Convey("ok", func() {
			r, err := s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "first"}, ActionID: "first-xxxx",
				BuildDependencies: []actions.Package{
					{Action: &core.Action{Name: "second"}, ActionID: "second-xxxx"},
					{
						Action: &core.Action{
							Name:     "third",
							Metadata: &core.Action_Metadata{Luciexe: &core.Action_Metadata_LUCIExe{StepName: "third-step"}},
						},
						BuildDependencies: []actions.Package{
							{Action: &core.Action{Name: "fourth"}, ActionID: "fourth-xxxx"},
						},
						ActionID: "third-xxxx",
					},
				},
			})
			So(err, ShouldBeNil)
			So(r.ID(), ShouldEqual, "first-xxxx")

			So(s.GetRoot("first-xxxx").ID(), ShouldEqual, "first-xxxx")
			So(s.GetRoot("second-xxxx").ID(), ShouldEqual, "first-xxxx")
			So(s.GetRoot("third-xxxx").ID(), ShouldEqual, "third-xxxx")
			So(s.GetRoot("fourth-xxxx").ID(), ShouldEqual, "third-xxxx")

			r, err = s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "third"}, ActionID: "third-xxxx",
			})
			So(err, ShouldBeNil)
			So(r.ID(), ShouldEqual, "third-xxxx")
		})

		Convey("conflict", func() {
			r, err := s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "first-1"}, ActionID: "first-1-xxxx",
				BuildDependencies: []actions.Package{
					{Action: &core.Action{Name: "second"}, ActionID: "second-xxxx"},
					{Action: &core.Action{
						Name:     "third",
						Metadata: &core.Action_Metadata{Luciexe: &core.Action_Metadata_LUCIExe{StepName: "third-step"}},
					}, ActionID: "third-xxxx"},
				},
			})
			So(err, ShouldBeNil)
			So(r.ID(), ShouldEqual, "first-1-xxxx")

			r, err = s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "first-2"}, ActionID: "first-2-xxxx",
				BuildDependencies: []actions.Package{
					{Action: &core.Action{
						Name:     "third",
						Metadata: &core.Action_Metadata{Luciexe: &core.Action_Metadata_LUCIExe{StepName: "third-step"}},
					}, ActionID: "third-xxxx"},
				},
			})
			So(err, ShouldBeNil)
			So(r.ID(), ShouldEqual, "first-2-xxxx")

			So(s.GetRoot("second-xxxx").ID(), ShouldEqual, "first-1-xxxx")
			So(s.GetRoot("third-xxxx").ID(), ShouldEqual, "third-xxxx")

			r, err = s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "first-3"}, ActionID: "first-3-xxxx",
				BuildDependencies: []actions.Package{
					{Action: &core.Action{Name: "second"}, ActionID: "second-xxxx"},
				},
			})
			So(r, ShouldBeNil)
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "must only belong to one root")

			r, err = s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "first-4"}, ActionID: "first-4-xxxx",
				RuntimeDependencies: []actions.Package{
					{Action: &core.Action{Name: "second"}, ActionID: "second-xxxx"},
				},
			})
			So(r, ShouldBeNil)
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "must only belong to one root")

			r, err = s.UpdateRoot(ctx, actions.Package{
				Action: &core.Action{Name: "second"}, ActionID: "second-xxxx",
			})
			So(r, ShouldBeNil)
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "top level package")
		})
	})
}

func TestStepUtilities(t *testing.T) {
	Convey("step utilities", t, func() {
		ctx := context.Background()

		Convey("runStepCommand", func() {
			self, err := os.Executable()
			So(err, ShouldBeNil)

			Convey("without iostream", func() {
				err := runStepCommand(ctx, &exec.Cmd{
					Path: self,
					Args: []string{self, "-help"},
				})
				So(err, ShouldBeNil)
			})
			Convey("with iostream", func() {
				b := bytes.NewBuffer(nil)
				err := runStepCommand(ctx, &exec.Cmd{
					Path:   self,
					Args:   []string{self, "-help"},
					Stdout: b,
					Stderr: b,
				})
				So(err, ShouldBeNil)
				So(b.String(), ShouldContainSubstring, "test")
			})
		})
	})
}
