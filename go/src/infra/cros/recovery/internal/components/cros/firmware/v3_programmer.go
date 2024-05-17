// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger"
)

// servodStateRecord holds state of servod before apply preparation of programmer.
type servodStateRecord struct {
	cmd string
	val interface{}
}

type v3Programmer struct {
	st     *servo.ServoType
	run    components.Runner
	servod components.Servod
	log    logger.Logger

	// Servod state before execution.
	servodState []*servodStateRecord
}

const (
	// Number of seconds for program EC/BIOS to time out.
	firmwareProgramTimeout = 30 * time.Minute

	// Tools and commands used for flashing EC.
	ecProgrammerToolName     = "flash_ec"
	ecProgrammerCmdGlob      = "flash_ec --chip=%s --image=%s --port=%d --verify --verbose"
	ecProgrammerITECmdGlob   = "flash_ec --chip=%s --image=%s --port=%d --verify --verbose --nouse_i2c_pseudo"
	ecProgrammerStm32CmdGlob = "flash_ec --chip=%s --image=%s --port=%d --bitbang_rate=57600 --verify --verbose"

	// Tools and commands used for flashing AP.
	apProgrammerToolName    = "futility"
	apProgrammerCmdGlob     = "futility update -i %s --servo_port=%d"
	apProgrammerWithGbbFlag = "--gbb_flags=%s"
	apProgrammerWithForce   = "--force"
)

// ProgramEC programs EC firmware to devices by servo.
func (p *v3Programmer) ProgramEC(ctx context.Context, fwBoard, imagePath string) error {
	if err := isFileExist(ctx, imagePath, p.run); err != nil {
		return errors.Annotate(err, "program ec").Err()
	}
	return p.programEC(ctx, fwBoard, imagePath)
}

// programEC programs EC firmware to devices by servo.
// Extracted for test purpose to avoid file present check.
func (p *v3Programmer) programEC(ctx context.Context, fwBoard, imagePath string) error {
	servoType, err := p.servoType(ctx)
	if err != nil {
		return errors.Annotate(err, "program ec").Err()
	}
	ecChip, err := p.ecChip(ctx)
	if err != nil {
		return errors.Annotate(err, "program ec").Err()
	}
	var cmd string
	if ecChip == "stm32" {
		cmd = fmt.Sprintf(ecProgrammerStm32CmdGlob, ecChip, imagePath, p.servod.Port())
	} else if strings.HasPrefix(ecChip, "it8") && !servoType.IsMicro() {
		// TODO(b/130037062): This special case can be removed when all labstations reach R126.
		cmd = fmt.Sprintf(ecProgrammerITECmdGlob, ecChip, imagePath, p.servod.Port())
	} else {
		cmd = fmt.Sprintf(ecProgrammerCmdGlob, ecChip, imagePath, p.servod.Port())
	}
	if fwBoard != "" {
		cmd += fmt.Sprintf(" --board=%s", fwBoard)
	}
	if p.ecUpdateRequiresApshutdown(ctx) {
		// Introduced due EC SW Sync race (b/269804618).
		cmd += " --try_apshutdown"
	}
	// Verify EC flash tool exists.
	if err := isToolPresent(ctx, ecProgrammerToolName, p.run); err != nil {
		return errors.Annotate(err, "program ec").Err()
	}
	out, err := p.run(ctx, firmwareProgramTimeout, cmd)
	p.log.Debugf("Program EC output: \n%s", out)
	return errors.Annotate(err, "program ec").Err()
}

// ProgramAP programs AP firmware to devices by servo.
//
// To set/update GBB flags please provide value in hex representation.
// E.g. 0x18 to set force boot in DEV-mode and allow to boot from USB-drive in DEV-mode.
// When force enabled, programmer will do force update (skip checking contents).
func (p *v3Programmer) ProgramAP(ctx context.Context, imagePath, gbbHex string, force bool) error {
	if err := isFileExist(ctx, imagePath, p.run); err != nil {
		return errors.Annotate(err, "program ap").Err()
	}
	return p.programAP(ctx, imagePath, gbbHex, force)
}

// programAP programs AP firmware to devices by servo.
// Extracted for test purpose to avoid file present check.
func (p *v3Programmer) programAP(ctx context.Context, imagePath, gbbHex string, force bool) error {
	if err := isToolPresent(ctx, apProgrammerToolName, p.run); err != nil {
		return errors.Annotate(err, "program ap").Err()
	}
	cmd := []string{
		fmt.Sprintf(apProgrammerCmdGlob, imagePath, p.servod.Port()),
	}
	if gbbHex != "" {
		cmd = append(cmd, fmt.Sprintf(apProgrammerWithGbbFlag, gbbHex))
	}
	if force {
		cmd = append(cmd, apProgrammerWithForce)
	}
	out, err := p.run(ctx, firmwareProgramTimeout, strings.Join(cmd, " "))
	p.log.Debugf("Program AP output:\n%s", out)
	return errors.Annotate(err, "program ap").Err()
}

// Prepare programmer for actions.
func (p *v3Programmer) Prepare(ctx context.Context) error {
	err := p.setServodState(ctx)
	return errors.Annotate(err, "prepare").Err()
}

func (p *v3Programmer) setServodState(ctx context.Context) error {
	p.log.Debugf("Set servod state to prepare programmer.")
	for _, s := range p.servodStateList() {
		sp := strings.Split(s, ":")
		if len(sp) != 2 {
			return errors.Reason("prepare servod state: state %q is incorrect", s).Err()
		}
		command := sp[0]
		val := sp[1]
		if cs, err := p.servod.Get(ctx, command); err != nil {
			return errors.Annotate(err, "prepare servod state: read servod state").Err()
		} else if val != cs.GetString_() {
			// If value is different then we need to save it so later we can restore it.
			r := &servodStateRecord{
				cmd: command,
				val: cs.GetString_(),
			}
			p.servodState = append(p.servodState, r)
		}
		if err := p.servod.Set(ctx, command, val); err != nil {
			return errors.Annotate(err, "prepare servod state: set servod state").Err()
		}
	}
	return nil
}

func (p *v3Programmer) restoreServodState(ctx context.Context) error {
	for _, s := range p.servodState {
		if err := p.servod.Set(ctx, s.cmd, s.val); err != nil {
			return errors.Annotate(err, "prepare servod state: set servod state").Err()
		}
	}
	return nil
}

func (p *v3Programmer) servodStateList() []string {
	if p.st.IsCCD() {
		return nil
	}
	return []string{
		"spi2_vref:pp3300", //Need verify as in some cases it can be pp1800
		"spi2_buf_en:on",
		"spi2_buf_on_flex_en:on",
		"spi_hold:off",
		"cold_reset:on",
		"usbpd_reset:on",
	}
}

// Close closes programming resources.
func (p *v3Programmer) Close(ctx context.Context) error {
	if err := p.restoreServodState(ctx); err != nil {
		return errors.Annotate(err, "close").Err()
	}
	return nil
}

// ecChip reads ec_chip from servod.
func (p *v3Programmer) ecChip(ctx context.Context) (string, error) {
	if ecChipI, err := p.servod.Get(ctx, "ec_chip"); err != nil {
		return "", errors.Annotate(err, "get ec_chip").Err()
	} else {
		return ecChipI.GetString_(), nil
	}
}

// servoType reads servo_type from servod.
func (p *v3Programmer) servoType(ctx context.Context) (*servo.ServoType, error) {
	if ecChipI, err := p.servod.Get(ctx, "servo_type"); err != nil {
		return nil, errors.Annotate(err, "get servo_type").Err()
	} else {
		return servo.NewServoType(ecChipI.GetString_()), nil
	}
}

// Controls map to tell if `--try_apshutdown` requires to use when EC update.
// Introduced due EC SW Sync race (b/269804618).
var ecUpdateRequiresApshutdownControls = []struct {
	checkControl  string // to tell presents of the target control.
	targetControl string //  to check if `--try_apshutdown` required to be used.
}{
	{"cpu_fw_spi", "cpu_fw_spi_depends_on_ec_fw"},
	{"ccd_cpu_fw_spi", "ccd_cpu_fw_spi_depends_on_ec_fw"},
}

// ecUpdateRequiresApshutdown checks if flash_ec needs apply '--try_apshutdown'.
// Both controls are checked because in CCD + debug header combo setups.
// Introduced due EC SW Sync race (b/269804618).
func (p *v3Programmer) ecUpdateRequiresApshutdown(ctx context.Context) bool {
	isSupported := func(presentCheckControl, targetControl string) bool {
		if err := p.servod.Has(ctx, presentCheckControl); err != nil {
			log.Debugf(ctx, "Target control %q is not present: %s", targetControl, err)
			return false
		}
		val, err := servo.GetString(ctx, p.servod, targetControl)
		if err != nil {
			log.Debugf(ctx, "Failed to read %q: %s", targetControl, err)
			return false
		}
		return val == "yes"
	}
	for _, controls := range ecUpdateRequiresApshutdownControls {
		if isSupported(controls.checkControl, controls.targetControl) {
			return true
		}
	}
	return false
}

// gbbToInt converts hex value to int.
//
// E.g. 0x18 to set force boot in DEV-mode and allow to boot from USB-drive in DEV-mode.
func gbbToInt(hex string) (int, error) {
	hex = strings.ToLower(hex)
	hexCut := strings.Replace(hex, "0x", "", -1)
	if v, err := strconv.ParseInt(hexCut, 16, 64); err != nil {
		return 0, errors.Annotate(err, "gbb to int").Err()
	} else {
		return int(v), nil
	}
}

// isFileExist checks is provided file exists.
func isFileExist(ctx context.Context, filepath string, run components.Runner) error {
	_, err := run(ctx, 30*time.Second, "test", "-f", filepath)
	return errors.Annotate(err, "if file exist: file %q does not exist", filepath).Err()
}

// isToolPresent checks if tool is installed on the host.
func isToolPresent(ctx context.Context, toolName string, run components.Runner) error {
	cmd := fmt.Sprintf("which %s", toolName)
	_, err := run(ctx, 30*time.Second, cmd)
	return errors.Annotate(err, "tool %s is not found", toolName).Err()
}
