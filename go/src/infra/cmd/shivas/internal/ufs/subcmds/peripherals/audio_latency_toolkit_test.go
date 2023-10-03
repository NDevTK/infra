// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package peripherals

import (
	"testing"
)

func Test_AudioLatencyToolkit(t *testing.T) {

	t.Parallel()

	t.Run("cleanAndValidateFlags", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			goodTests := []struct {
				cmd *manageAudioLatencyToolkitCmd
			}{
				{cmd: &manageAudioLatencyToolkitCmd{dutName: "d", version: "4.1"}},
			}
			for _, tt := range goodTests {
				err := tt.cmd.cleanAndValidateFlags()
				if err != nil {
					t.Errorf("cleanAndValidateFlags got error: %v; want: nil", err)
					continue
				}
			}
		})
		t.Run("invalid", func(t *testing.T) {
			errTests := []struct {
				cmd  *manageAudioLatencyToolkitCmd
				want []string
			}{
				{
					cmd:  &manageAudioLatencyToolkitCmd{version: "4.1"},
					want: []string{errDUTMissing},
				},
				{
					cmd:  &manageAudioLatencyToolkitCmd{dutName: " ", version: "4.1"},
					want: []string{errDUTMissing},
				},
			}

			for _, tt := range errTests {
				err := tt.cmd.cleanAndValidateFlags()
				if err == nil {
					t.Errorf("cleanAndValidateFlags = nil; want errors: %v", tt.want)
					continue
				}
			}
		})
	})

	t.Run("add action", func(t *testing.T) {
		// Valid case
		c := &manageAudioLatencyToolkitCmd{
			dutName: "d",
			version: "4.1",
			mode:    actionAdd,
		}

		t.Run("valid createAudioLatencyToolkit", func(t *testing.T) {
			res, err := c.createAudioLatencyToolkit()
			if err != nil {
				t.Errorf("unable to create Audio Latency Toolkit: %v; want nil", err)
			}

			if res.GetVersion() != "4.1" {
				t.Errorf("invalid version in Audio Latency Toolkit: %s; want 4.1", res.GetVersion())
			}
		})

		t.Run("valid runAudioLatencyToolkitAction", func(t *testing.T) {
			res, err := c.runAudioLatencyToolkitAction()
			if err != nil {
				t.Errorf("unable to create Audio Latency Toolkit: %v; want nil", err)
			}
			if res.GetVersion() != "4.1" {
				t.Errorf("invalid version in Audio Latency Toolkit: %s; want 4.1", res.GetVersion())
			}
		})
	})

	t.Run("delete action", func(t *testing.T) {
		// Valid case
		c := &manageAudioLatencyToolkitCmd{
			dutName: "d",
			version: "4.1",
			mode:    actionDelete,
		}

		t.Run("valid runAudioLatencyToolkitAction", func(t *testing.T) {
			res, err := c.runAudioLatencyToolkitAction()
			if err != nil {
				t.Errorf("unable to delete Audio Latency Toolkit: %v; want nil", err)
			}
			if res != nil {
				t.Errorf("Audio Latency Toolkit not deleted got: %v; want nil", res)
			}
		})
	})

	t.Run("replace action", func(t *testing.T) {
		c := &manageAudioLatencyToolkitCmd{
			dutName: "d",
			version: "4.1",
			mode:    actionReplace,
		}

		t.Run("invalid action runAudioLatencyToolkitAction", func(t *testing.T) {
			if _, err := c.runAudioLatencyToolkitAction(); err == nil {
				t.Errorf("does not receive error for replace action: %v; want unknown action error", err)
			}
		})
	})

}
