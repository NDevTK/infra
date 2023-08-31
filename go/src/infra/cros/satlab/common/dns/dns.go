package dns

import (
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/executor"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"
)

type DNSCommand struct {
	ExecCommander executor.IExecCommander
}

func NewDNSCommand() *DNSCommand {
	return &DNSCommand{
		ExecCommander: &executor.ExecCommander{},
	}
}

// readContents gets the content of a DNS file.
// If the DNS file does not exist, replace it with an empty container.
func (d *DNSCommand) ReadContents() (string, error) {
	// Defensively touch the file if it does not already exist.
	// See b/199796469 for details.
	args := []string{
		paths.DockerPath,
		"exec",
		"dns",
		"touch",
		"/etc/dut_hosts/hosts",
	}
	if err := d.ExecCommander.Run(exec.Command(args[0], args[1:]...)); err != nil {
		return "", errors.Annotate(err, "defensively touch dns file").Err()
	}
	args = []string{
		paths.DockerPath,
		"exec",
		"dns",
		"/bin/cat",
		"/etc/dut_hosts/hosts",
	}
	out, err := d.ExecCommander.Exec(exec.Command(args[0], args[1:]...))
	return strings.TrimRight(
			string(out),
			"\n\t",
		), errors.Annotate(err, "get dns file content").
			Err()
}

// ReadHostsIP read the hosts file to get the IP Host Mapping
// Or Host IP Mapping
func (d *DNSCommand) ReadHostsIP(useIPAsKey bool) (map[string]string, error) {
	res := map[string]string{}
	rawData, err := d.ReadContents()

	if err != nil {
		return res, nil
	}

	list := strings.Split(rawData, "\n")

	for _, row := range list {
		r := strings.Split(row, "\t")
		// We only handle vaild data
		// e.g. <ip>\t<hostname>
		if len(r) == 2 {
			if useIPAsKey {
				res[r[0]] = r[1]
			} else {
				res[r[1]] = r[0]
			}
		}
	}

	return res, nil
}
