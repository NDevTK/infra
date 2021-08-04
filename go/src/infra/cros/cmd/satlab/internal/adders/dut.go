package adders

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/paths"
)

// DUT contains all the information necessary to add a DUT.
type DUT struct {
	Namespace  string
	Zone       string
	Host       string
	Servo      string
	ShivasArgs map[string][]string
}

// Run adds a DUT if it does not already exist.
func (d *DUT) Run() error {
	args := (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasPath, "get", "dut"},
		Flags: map[string][]string{
			"namespace": {d.Namespace},
			"zone":      {d.Zone},
		},
		PositionalArgs: []string{d.Host},
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add dut if applicable: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	dutMsgBytes, err := command.Output()
	dutMsg := commands.TrimOutput(dutMsgBytes)
	if err != nil {
		return errors.Annotate(err, "add dut if applicable: running %s", strings.Join(args, " ")).Err()
	}
	if len(dutMsg) == 0 {
		fmt.Fprintf(os.Stderr, "Adding DUT\n")

		flags := make(map[string][]string)
		for k, v := range d.ShivasArgs {
			flags[k] = v
		}

		flags["name"] = []string{d.Host}
		// This flag must have the form labstation:port.
		// Do not validate this flag here since we don't want to potentially drift
		// out of sync with the format that shivas expects.
		// TODO(gregorynisbet): Consider pre-populating it.
		flags["servo"] = []string{d.Servo}

		// TODO(gregorynisbet): Consider a different strategy for tracking flags
		// that cannot be passed to shivas add dut.
		args := (&commands.CommandWithFlags{
			Commands: []string{paths.ShivasPath, "add", "dut"},
			Flags:    flags,
		}).ToCommand()
		fmt.Fprintf(os.Stderr, "Add dut if applicable: run %s\n", args)
		command := exec.Command(args[0], args[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return errors.Annotate(
				err,
				fmt.Sprintf(
					"add dut if applicable: running %s",
					strings.Join(args, " "),
				),
			).Err()
		}
	} else {
		fmt.Fprintf(os.Stderr, "DUT already added\n")
	}
	return nil
}
