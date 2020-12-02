package rtd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/config/go/api/test/rtd/v1"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

const volumeContainerDir = "/vol"

var (
	containerRunning = false
	containerHash    = ""
	volumeHostDir    = ""
	volumeName       = ""
)

// StartRTDContainer starts an RTD container, possibly a totally fake one, and
// returns once it's running.
func StartRTDContainer(ctx context.Context, imageUrl string) error {
	if containerRunning {
		return fmt.Errorf("container already started; can't start another one")
	}
	logging.Infof(ctx, "Starting RTD container")
	var err error
	if volumeHostDir, err = ioutil.TempDir(os.TempDir(), "rtd-volume"); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	var stdoutBuf, stderrBuf bytes.Buffer
	if err = dockerImpl.run(ctx, &stdoutBuf, &stderrBuf, "pull", imageUrl); err != nil {
		return errors.New(stderrBuf.String())
	}
	if err = dockerImpl.run(ctx, &stdoutBuf, &stderrBuf, "volume", "create", "--driver", "local", "--opt", "type=none", "--opt", "device="+volumeHostDir, "--opt", "o=bind"); err != nil {
		return errors.New(stderrBuf.String())
	}
	volumeName = strings.TrimSpace(stdoutBuf.String())
	logging.Infof(ctx, "Created Docker volume %s", volumeName)
	if err := dockerImpl.run(ctx, &stdoutBuf, &stderrBuf, "run", "--network", "host", "-t", "--mount", "source="+volumeName+",target="+volumeContainerDir, "-d", imageUrl); err != nil {
		return errors.New(stderrBuf.String())
	}
	containerHash = strings.TrimSpace(stdoutBuf.String())
	logging.Infof(ctx, "RTD container started with hash %v", containerHash)
	containerRunning = true
	return nil
}

// StopRTDContainer stops a running RTD container.
func StopRTDContainer(ctx context.Context) error {
	if !containerRunning {
		return fmt.Errorf("container isn't running; nothing to stop")
	}
	logging.Infof(ctx, "Stopping RTD container")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	var stdoutBuf, stderrBuf bytes.Buffer
	if err := dockerImpl.run(ctx, &stdoutBuf, &stderrBuf, "stop", containerHash); err != nil {
		return errors.New(stderrBuf.String())
	}
	logging.Infof(ctx, "Stopped the RTD container")
	containerRunning = false
	containerHash = ""
	return nil
}

// Invoke runs an RTD Invocation against the running RTD container.
func Invoke(ctx context.Context, progressSinkPort, tlsPort int32, rtdCmd string) error {
	if !containerRunning {
		return fmt.Errorf("container hasn't been started yet; can't invoke")
	}
	// TODO: needs more work
	i := &rtd.Invocation{
		ProgressSinkClientConfig: &rtd.ProgressSinkClientConfig{
			Port: progressSinkPort,
		},
		TestLabServicesConfig: &rtd.TLSClientConfig{
			TlsPort:    tlsPort,
			TlsAddress: "127.0.0.1",
		},
		Duts: []*rtd.DUT{
			{
				TlsDutName: "my-little-dutty",
			},
		},
		Requests: []*rtd.Request{
			{
				Name: "request_dummy-pass",
				Test: "remoteTestDrivers/tnull/tests/dummy-pass",
			},
		},
	}
	invocationFile, err := writeInvocationToFile(ctx, i)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "TODO: `docker exec` against the container, with --input %v", invocationFile)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	var stdoutBuf, stderrBuf bytes.Buffer
	dockerCmd := []string{"exec", containerHash}
	dockerCmd = append(dockerCmd, strings.Fields(strings.Trim(rtdCmd, "\""))...)
	dockerCmd = append(dockerCmd, "--input", invocationFile)
	if err := dockerImpl.run(ctx, &stdoutBuf, &stderrBuf, dockerCmd...); err != nil {
		return errors.New(stderrBuf.String())
	}
	return nil
}

func writeInvocationToFile(ctx context.Context, i *rtd.Invocation) (string, error) {
	b, err := proto.Marshal(i)
	if err != nil {
		return "", err
	}
	filename := "invocation.binaryproto"
	localFile := path.Join(volumeHostDir, filename)
	remoteFile := path.Join(volumeContainerDir, filename)
	if err = ioutil.WriteFile(localFile, b, 0664); err != nil {
		return "", err
	}
	marsh := jsonpb.Marshaler{EmitDefaults: true, Indent: "  "}
	strForm, err := marsh.MarshalToString(i)
	if err != nil {
		return "", err
	}
	logging.Infof(ctx, "Wrote RTD's input Invocation binaryproto to %v", localFile)
	logging.Infof(ctx, "Contents of this Invocation message in jsonpb form are:\n%v", strForm)
	return remoteFile, nil
}
