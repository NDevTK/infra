// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cellular contains utilities for repairing cellular DUTs.
package cellular

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/internal/retry"
)

const (
	modemManagerJob           = "modemmanager"
	detectCmd                 = "mmcli -m a -J"
	mmcliCliPresentCmd        = "which mmcli"
	modemManagerJobPresentCmd = "initctl status modemmanager"
	restartModemManagerCmd    = "restart modemmanager"
	getSignalStrengthCmd      = "mmcli -m a --signal-get -J"
	startModemManagerCmd      = "start modemmanager"
	shillInterface            = "org.chromium.flimflam"
	getCellularServiceCmd     = "gdbus call --system --dest=org.chromium.flimflam" +
		" -o / -m org.chromium.flimflam.Manager.FindMatchingService" +
		" \"{'Type': 'cellular'}\" | cut -d\"'\" -f2"
)

// HasModemManagerCLI returns true if mmcli is present on the DUT.
func HasModemManagerCLI(ctx context.Context, runner components.Runner, timeout time.Duration) bool {
	if _, err := runner(ctx, timeout, mmcliCliPresentCmd); err != nil {
		return false
	}
	return true
}

// HasModemManagerJob returns true if modemmanager job is present on the DUT.
func HasModemManagerJob(ctx context.Context, runner components.Runner, timeout time.Duration) bool {
	if _, err := runner(ctx, timeout, modemManagerJobPresentCmd); err != nil {
		return false
	}
	return true
}

// StartModemManager starts modemmanager via upstart.
func StartModemManager(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	if _, err := runner(ctx, timeout, startModemManagerCmd); err != nil {
		return errors.Annotate(err, "start modemmanager").Err()
	}
	return nil
}

// RestartModemManager restarts modemmanager via upstart.
func RestartModemManager(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	if _, err := runner(ctx, timeout, restartModemManagerCmd); err != nil {
		return errors.Annotate(err, "restart modemmanager").Err()
	}
	return nil
}

// ConnectToDefaultService attempts a simple connection to the default cellular service.
func ConnectToDefaultService(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	info, err := WaitForModemInfo(ctx, runner, 15*time.Second)
	if err != nil {
		return errors.Annotate(err, "connect to default service: get modem info").Err()
	}

	// skip if already in connected state
	if strings.EqualFold(info.GetState(), string(ModemStateConnected)) {
		log.Infof(ctx, "connect to default service: modem is already connected to service")
		return nil
	}

	// don't attempt to connect if in "connecting" state
	if !strings.EqualFold(info.GetState(), string(ModemStateConnecting)) {
		serviceName, err := runner(ctx, 5*time.Second, getCellularServiceCmd)
		if err != nil {
			return errors.Annotate(err, "connect to default service: get service name").Err()
		}

		connectCmd := fmt.Sprintf("dbus-send --system --fixed --print-reply --dest=%s %s %s.Service.Connect", shillInterface, serviceName, shillInterface)
		if _, err := runner(ctx, 30*time.Second, connectCmd); err != nil {
			return errors.Annotate(err, "connect to default service").Err()
		}
	}

	if err := WaitForModemState(ctx, runner, timeout, ModemStateConnected); err != nil {
		return errors.Annotate(err, "connect to default service: modem never entered connected state").Err()
	}
	return nil
}

// WaitForModemManager waits for the modemmanager job to be running via upstart.
func WaitForModemManager(ctx context.Context, runner components.Runner, timeout time.Duration) error {
	cmd := fmt.Sprintf("status %s", modemManagerJob)
	return retry.WithTimeout(ctx, time.Second, timeout, func() error {
		if output, err := runner(ctx, 5*time.Second, cmd); err != nil {
			return errors.Annotate(err, "get modemmanager status").Err()
		} else if !strings.Contains(output, "start/running") {
			return errors.Reason("modemmanager not running").Err()
		}
		return nil
	}, "wait for modemmanager")
}

// WaitForModemState polls for the modem to enter a specific state.
func WaitForModemState(ctx context.Context, runner components.Runner, timeout time.Duration, state ModemState) error {
	predicate := func(m *ModemInfo) error {
		if m.GetState() == "" {
			return errors.Reason("modem state is empty").Err()
		}

		if !strings.EqualFold(m.GetState(), string(state)) {
			return errors.Reason("modem state not equal to %s", state).Err()
		}
		return nil
	}

	if _, err := WaitForModemInfo(ctx, runner, timeout, predicate); err != nil {
		return errors.Annotate(err, "wait for modem state: wait for modem to enter requested state").Err()
	}

	return nil
}

// ModemInfo is a simplified version of the JSON output from ModemManager to get the modem connection state information.
type ModemInfo struct {
	Modem *struct {
		G3PP *struct {
			Imei string `json:"imei,omitempty"`
		} `json:"3gpp,omitempty"`
		Generic *struct {
			ActiveSIMSlot string   `json:"primary-sim-slot,omitempty"`
			State         string   `json:"state,omitempty"`
			SIM           string   `json:"sim,omitempty"`
			SIMSlots      []string `json:"sim-slots,omitempty"`
			OwnNumbers    []string `json:"own-numbers,omitempty"`
		} `json:"generic,omitempty"`
	} `json:"modem,omitempty"`
}

// ModemState represents a valid cellular modem state.
type ModemState string

const (
	// ModemStateConnected represents a modem connected to a cellular network.
	ModemStateConnected ModemState = "CONNECTED"
	// ModemStateConnecting represents a modem in the process of connecting to a cellular network.
	ModemStateConnecting ModemState = "CONNECTING"
)

// ActiveSIMSlot returns the currently active modem SIM slot.
func (m *ModemInfo) ActiveSIMSlot() int32 {
	if m == nil || m.Modem == nil || m.Modem.Generic == nil {
		return 0
	}

	if m.Modem.Generic.ActiveSIMSlot == "" || m.Modem.Generic.ActiveSIMSlot == "--" {
		return 0
	}

	i, err := strconv.ParseInt(m.Modem.Generic.ActiveSIMSlot, 10, 32)
	if err != nil {
		return 0
	}
	return int32(i)
}

// ActiveSIMID returns the ID of the active SIM slot.
func (m *ModemInfo) ActiveSIMID() string {
	if m == nil || m.Modem == nil || m.Modem.Generic == nil {
		return ""
	}

	if m.Modem.Generic.SIM == "" || m.Modem.Generic.SIM == "--" {
		return ""
	}

	simPath := strings.Split(m.Modem.Generic.SIM, "/")
	return simPath[len(simPath)-1]
}

// OwnNumber returns the modem's current phone number.
func (m *ModemInfo) OwnNumber() string {
	if m == nil || m.Modem == nil || m.Modem.Generic == nil {
		return ""
	}

	for _, number := range m.Modem.Generic.OwnNumbers {
		// Strip the country code from the phone number e.g. +1
		if len(number) > 10 {
			return number[len(number)-10:]
		}
		if len(number) == 10 {
			return number
		}
	}
	return ""
}

// SIMSlotCount returns the number of SIM slots available on the device.
func (m *ModemInfo) SIMSlotCount() int32 {
	if m == nil || m.Modem == nil || m.Modem.Generic == nil {
		return 0
	}
	return int32(len(m.Modem.Generic.SIMSlots))
}

// GetState returns the modems state as reported by ModemManager.
func (m *ModemInfo) GetState() string {
	if m == nil || m.Modem == nil || m.Modem.Generic == nil {
		return ""
	}
	return m.Modem.Generic.State
}

func (m *ModemInfo) GetImei() string {
	// ModemManager may replace missing fields with "--"
	if m == nil || m.Modem == nil || m.Modem.G3PP == nil || strings.EqualFold(m.Modem.G3PP.Imei, "--") {
		return ""
	}
	return m.Modem.G3PP.Imei
}

// ModemPredicate returns an error if the modem is not in the correct state.
type ModemPredicate func(m *ModemInfo) error

// WaitForModemInfo polls for a modem to appear on the DUT, which can take up to two minutes on reboot.
func WaitForModemInfo(ctx context.Context, runner components.Runner, timeout time.Duration, predicates ...ModemPredicate) (*ModemInfo, error) {
	var info *ModemInfo
	if err := retry.WithTimeout(ctx, time.Second, timeout, func() error {
		output, err := runner(ctx, 5*time.Second, detectCmd)
		if err != nil {
			return errors.Annotate(err, "call mmcli").Err()
		}

		// Note: info is defined in outer scope as retry.WithTimeout only allows returning errors.
		info, err = parseModemInfo(ctx, output)
		if err != nil {
			return errors.Annotate(err, "parse mmcli response").Err()
		}

		if info == nil || info.Modem == nil {
			return errors.Reason("no modem found on DUT").Err()
		}

		// Wait for any additional state requirements.
		for _, predicate := range predicates {
			if err := predicate(info); err != nil {
				return errors.Annotate(err, "failed predicate").Err()
			}
		}

		return nil
	}, "wait for modem"); err != nil {
		return nil, errors.Annotate(err, "wait for modem info: wait for ModemManager to export modem").Err()
	}

	return info, nil
}

// parseModemInfo unmarshals the modem properties json output from mmcli.
func parseModemInfo(ctx context.Context, output string) (*ModemInfo, error) {
	info := &ModemInfo{}
	if err := json.Unmarshal([]byte(output), info); err != nil {
		return nil, err
	}
	return info, nil
}

// signalInfo is a simplified version of the JSON output from ModemManager to get the modem signal strength information
type signalInfo struct {
	Modem *struct {
		Signal *struct {
			FiveG *struct {
				RSRP string `json:"rsrp,omitempty"`
				RSSI string `json:"rssi,omitempty"`
				SNR  string `json:"snr,omitempty"`
			} `json:"5g,omitempty"`
			LTE *struct {
				RSRP string `json:"rsrp,omitempty"`
				RSSI string `json:"rssi,omitempty"`
				SNR  string `json:"snr,omitempty"`
			} `json:"lte,omitempty"`
		} `json:"signal,omitempty"`
	} `json:"modem,omitempty"`
}

// NetworkTechnology represents a cellular network technology.
type NetworkTechnology string

const (
	// NetworkTechnologyLTE represents an LTE cellular network.
	NetworkTechnologyLTE NetworkTechnology = "LTE"
	// NetworkTechnology5G represents a 5G cellular network.
	NetworkTechnology5G NetworkTechnology = "5G"
)

// SignalStrength represents a generic cellular signal measurement. Different modems may
// report all or only some of the possible measurements, unavailable measurements are set
// to nil.
type SignalStrength struct {
	RSRP       *float64
	RSSI       *float64
	SNR        *float64
	Technology NetworkTechnology
}

func (s SignalStrength) HasValue() bool {
	return s.RSRP != nil || s.RSSI != nil || s.SNR != nil
}

// GetSignalStrength fetches the available cellular signals and returns their strengths.
// Note: Multiple signal technologies may be available at one time (i.e. 5G and LTE) in which case
// all available will be returned.
func GetSignalStrength(ctx context.Context, runner components.Runner, timeout time.Duration) ([]SignalStrength, error) {
	signalStrength := make([]SignalStrength, 0)
	if err := retry.WithTimeout(ctx, time.Second, timeout, func() error {
		output, err := runner(ctx, 5*time.Second, getSignalStrengthCmd)
		if err != nil {
			return errors.Annotate(err, "query signal").Err()
		}

		info, err := parseSignalInfo(ctx, output)
		if err != nil {
			return errors.Annotate(err, "parse signal response").Err()
		}

		if info == nil || info.Modem == nil || info.Modem.Signal == nil {
			return errors.Reason("no signal info found on DUT").Err()
		}

		if info.Modem.Signal.FiveG != nil {
			// CLI may return "--" or empty strings for missing signals so we need
			// to verify by attempting to parse.
			strength := SignalStrength{}
			if rsrp, err := strconv.ParseFloat(info.Modem.Signal.FiveG.RSRP, 32); err == nil {
				strength.RSRP = &rsrp
			}
			if rssi, err := strconv.ParseFloat(info.Modem.Signal.FiveG.RSSI, 32); err == nil {
				strength.RSSI = &rssi
			}
			if snr, err := strconv.ParseFloat(info.Modem.Signal.FiveG.SNR, 32); err == nil {
				strength.SNR = &snr
			}
			if strength.HasValue() {
				strength.Technology = NetworkTechnology5G
				signalStrength = append(signalStrength, strength)
			}
		}

		if info.Modem.Signal.LTE != nil {
			strength := SignalStrength{}
			if rsrp, err := strconv.ParseFloat(info.Modem.Signal.LTE.RSRP, 32); err == nil {
				strength.RSRP = &rsrp
			}
			if rssi, err := strconv.ParseFloat(info.Modem.Signal.LTE.RSSI, 32); err == nil {
				strength.RSSI = &rssi
			}
			if snr, err := strconv.ParseFloat(info.Modem.Signal.LTE.SNR, 32); err == nil {
				strength.SNR = &snr
			}
			if strength.HasValue() {
				strength.Technology = NetworkTechnologyLTE
				signalStrength = append(signalStrength, strength)
			}
		}

		// if any available signals were found
		if len(signalStrength) > 0 {
			return nil
		}
		return errors.Reason("no signal info found on DUT").Err()
	}, "wait for modem"); err != nil {
		return nil, errors.Annotate(err, "wait for modem info: wait for ModemManager to export modem").Err()
	}

	return signalStrength, nil
}

// parseSignalInfo unmarshals the modem signal properties json output from mmcli.
func parseSignalInfo(ctx context.Context, output string) (*signalInfo, error) {
	info := &signalInfo{}
	if err := json.Unmarshal([]byte(output), info); err != nil {
		return nil, err
	}
	return info, nil
}
