// Copyright 2022 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package localtlw provides local implementation of TLW Access.
package localtlw

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.chromium.org/chromiumos/config/go/api/test/xmlrpc"
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/docker"
	"infra/cros/recovery/internal/localtlw/localproxy"
	"infra/cros/recovery/internal/localtlw/servod"
	tlw_xmlrpc "infra/cros/recovery/internal/localtlw/xmlrpc"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// InitServod initiates servod daemon on servo-host.
func (c *tlwClient) InitServod(ctx context.Context, req *tlw.InitServodRequest) error {
	dut, err := c.getDevice(ctx, req.Resource)
	if err != nil {
		return errors.Annotate(err, "init servod %q", req.Resource).Err()
	}
	chromeos := dut.GetChromeos()
	if chromeos.GetServo().GetName() == "" {
		return errors.Reason("init servod %q: servo is not found", req.Resource).Err()
	}
	if isServodContainer(dut) {
		err := c.prepareServodContainer(ctx, dut, req.Options, !req.GetNoServod())
		return errors.Annotate(err, "init servod %q", req.Resource).Err()
	}
	if !req.GetNoServod() {
		options := req.GetOptions()
		// Always override port by port specified in data.
		options.ServodPort = chromeos.GetServo().GetServodPort()
		if err := servod.StartServod(ctx, &servod.StartServodRequest{
			Host:    localproxy.BuildAddr(chromeos.GetServo().GetName()),
			SSHPool: c.sshPool,
			Options: options,
		}); err != nil {
			return errors.Annotate(err, "init servod %q", req.Resource).Err()
		}
	} else {
		// Just try to stop servod if it is running.
		if err := c.stopServodOnHardwareHost(ctx, chromeos); err != nil {
			log.Debugf(ctx, "(Not critical) Fail to stop servod as requested to prepare servo-host without servod daemon: %s", err)
		}
	}
	return nil
}

// dockerServodImageName provides image for servod when use container.
func dockerServodImageName() string {
	label := getEnv("SERVOD_CONTAINER_LABEL", "release")
	registry := getEnv("REGISTRY_URI", "us-docker.pkg.dev/chromeos-partner-moblab/common-core")
	return fmt.Sprintf("%s/servod:%s", registry, label)
}

// getEnv retrieves the value of the environment variable named by the key.
// If retrieved value is empty return default value.
func getEnv(key, defaultvalue string) string {
	if key != "" {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return defaultvalue
}

// createServodContainerArgs creates default args for servodContainer.
func createServodContainerArgs(detached bool, envVar, cmd []string) *docker.ContainerArgs {
	return &docker.ContainerArgs{
		Detached:   detached,
		EnvVar:     envVar,
		ImageName:  dockerServodImageName(),
		Network:    defaultDockerNetwork(),
		Volumes:    []string{"/dev:/dev"},
		Privileged: true,
		Exec:       cmd,
	}
}

// prepareServodContainer prepares servod container and start servod if required.
func (c *tlwClient) prepareServodContainer(ctx context.Context, dut *tlw.Dut, o *tlw.ServodOptions, startServod bool) error {
	containerName := servoContainerName(dut)
	d, err := c.dockerClient(ctx)
	if err != nil {
		return errors.Annotate(err, "start servod container").Err()
	}
	// Print all containers to know if something wrong.
	d.PrintAll(ctx)
	if up, err := d.IsUp(ctx, containerName); err != nil {
		return errors.Annotate(err, "start servod container").Err()
	} else if up {
		// Wait 2 second to detect servod as docker already running is expected to have it or not.
		waitTime := 2
		isUpErr := dockerVerifyServodDaemonIsUp(ctx, d, containerName, o.ServodPort, waitTime)
		needStopContainer := false
		if startServod && isUpErr != nil {
			// Container running without servod daemon. We need to stop it and start one with servod.
			needStopContainer = true
		} else if !startServod && isUpErr == nil {
			// Container running with servod daemon. We need to stop it and start one without servod.
			needStopContainer = true
		}
		if needStopContainer {
			d.Remove(ctx, containerName, true)
			log.Debugf(ctx, "Stopped container %q as running with incorrect service.", containerName)
		} else {
			log.Debugf(ctx, "Servod container %s is already up!", containerName)
			return nil
		}
	}
	envVar := servod.GenerateParams(o)
	var containerStartArgs []string
	if startServod {
		containerStartArgs = []string{"bash", "/start_servod.sh"}
	} else {
		containerStartArgs = []string{"tail", "-f", "/dev/null"}
	}
	containerArgs := createServodContainerArgs(true, envVar, containerStartArgs)
	// We always need to call pull before start as it will verify that image we used is latest.
	// If pull is missed the image will be used from local docker cache.
	// Image is small is expected to be download in less 1 minute but for safety we set 5.
	// TODO(gregorynisbet): Need collect info how long it takes in field.
	if err := d.Pull(ctx, containerArgs.ImageName, 5*time.Minute); err != nil {
		return errors.Annotate(err, "start servod container").Err()
	}
	// Servod expected to start in less 1 minutes and we set 2 in case there is any issue is exist.
	res, err := d.Start(ctx, containerName, containerArgs, 2*time.Minute)
	if err != nil {
		return errors.Annotate(err, "start servod container").Err()
	}
	log.Debugf(ctx, "Container started with id:%s\n with errout: %s", res.Stdout, res.Stderr)
	if startServod {
		// Wait 3 seconds as sometimes container is not fully initialized and fail
		// when start ing working with servod or tooling.
		// TODO(otabek): Move to servod-container wrapper.
		time.Sleep(3 * time.Second)
		// Waiting to finish servod initialization.
		if err := dockerVerifyServodDaemonIsUp(ctx, d, containerName, o.ServodPort, 60); err != nil {
			return errors.Annotate(err, "start servod container").Err()
		}
	}
	log.Debugf(ctx, "Servod container %s started and up!", containerName)
	return nil
}

// dockerVerifyServodDaemonIsUp verifies servod is running on servod daemon is up in container.
func dockerVerifyServodDaemonIsUp(ctx context.Context, dc docker.Client, containerName string, servodPort int32, waitTime int) error {
	eReq := &docker.ExecRequest{
		Timeout: 2 * time.Minute,
		Cmd:     []string{"servodtool", "instance", "wait-for-active", "-p", fmt.Sprintf("%d", servodPort), "--timeout", fmt.Sprintf("%d", waitTime)},
	}
	res, err := dc.Exec(ctx, containerName, eReq)
	// When the wait time is less request timeout, the cmd can finish with nil error
	// exitcode can be 1 if the wait time is exceeded.
	if res != nil && res.ExitCode != 0 {
		log.Debugf(ctx, "servodtool did not response before %s", waitTime)
		return errors.Reason(res.Stderr).Err()
	}
	return errors.Annotate(err, "docker verify servod daemon is up").Err()
}

// defaultDockerNetwork provides network in which docker need to run.
func defaultDockerNetwork() string {
	network := os.Getenv("DOCKER_DEFAULT_NETWORK")
	// If not provided then use host network.
	if network == "" {
		network = "host"
	}
	return network
}

// StopServod stops servod daemon on servo-host.
func (c *tlwClient) StopServod(ctx context.Context, resourceName string) error {
	dut, err := c.getDevice(ctx, resourceName)
	if err != nil {
		return errors.Annotate(err, "stop servod %q", resourceName).Err()
	}
	chromeos := dut.GetChromeos()
	if chromeos.GetServo().GetName() == "" {
		return errors.Reason("stop servod %q: servo is not found", resourceName).Err()
	}
	if isServodContainer(dut) {
		if d, err := c.dockerClient(ctx); err != nil {
			return errors.Annotate(err, "stop servod %q", resourceName).Err()
		} else {
			err := d.Remove(ctx, servoContainerName(dut), true)
			return errors.Annotate(err, "stop servod %q", resourceName).Err()
		}
	}
	if err := c.stopServodOnHardwareHost(ctx, chromeos); err != nil {
		return errors.Annotate(err, "stop servod %q", resourceName).Err()
	}
	return nil
}

// stopServodOnHardwareHost stops servod on labstation and servo_v3.
func (c *tlwClient) stopServodOnHardwareHost(ctx context.Context, chromeos *tlw.ChromeOS) error {
	err := servod.StopServod(ctx, &servod.StopServodRequest{
		Host:    localproxy.BuildAddr(chromeos.GetServo().GetName()),
		SSHPool: c.sshPool,
		Options: &tlw.ServodOptions{
			ServodPort: chromeos.GetServo().GetServodPort(),
		},
	})
	return errors.Annotate(err, "stop servod on hardware host").Err()
}

// CallServod executes a command on servod related to resource name.
// Commands will be run against servod on servo-host.
func (c *tlwClient) CallServod(ctx context.Context, req *tlw.CallServodRequest) *tlw.CallServodResponse {
	dut, err := c.getDevice(ctx, req.Resource)
	if err != nil {
		return generateFailCallServodResponse(ctx, req.GetResource(), err)
	}
	chromeos := dut.GetChromeos()
	if chromeos.GetServo().GetName() == "" {
		return generateFailCallServodResponse(ctx, req.GetResource(), errors.Reason("call servod %q: servo not found", req.GetResource()).Err())
	}
	// For container connect to the container as it running on the same host.
	if isServodContainer(dut) {
		return c.callServodOnContainer(ctx, req, dut)
	}
	return c.callServodOnHost(ctx, req, dut)
}

// generateFailCallServodResponse creates response for fail cases when call servod.
func generateFailCallServodResponse(ctx context.Context, resource string, err error) *tlw.CallServodResponse {
	log.Debugf(ctx, "Call servod fail with %s", err)
	return &tlw.CallServodResponse{
		Value: &xmlrpc.Value{
			ScalarOneof: &xmlrpc.Value_String_{
				String_: fmt.Sprintf("call servod %q: %s", resource, err),
			},
		},
		Fault: true,
	}
}

// callServodOnContainer calls servod running on servod-container.
func (c *tlwClient) callServodOnContainer(ctx context.Context, req *tlw.CallServodRequest, dut *tlw.Dut) *tlw.CallServodResponse {
	d, err := c.dockerClient(ctx)
	if err != nil {
		return generateFailCallServodResponse(ctx, req.GetResource(), err)
	}
	addr, err := d.IPAddress(ctx, servoContainerName(dut))
	if err != nil {
		return generateFailCallServodResponse(ctx, req.GetResource(), err)
	}
	timeout := req.GetTimeout().AsDuration()
	rpc := tlw_xmlrpc.New(addr, int(dut.GetChromeos().GetServo().GetServodPort()))
	if val, err := servod.Call(ctx, rpc, timeout, req.Method, req.Args); err != nil {
		return generateFailCallServodResponse(ctx, req.GetResource(), err)
	} else {
		return &tlw.CallServodResponse{
			Value: val,
			Fault: false,
		}
	}
}

// callServodOnHost calls servod running on physical host.
func (c *tlwClient) callServodOnHost(ctx context.Context, req *tlw.CallServodRequest, dut *tlw.Dut) *tlw.CallServodResponse {
	servoHost := dut.GetChromeos().GetServo()
	if servoHost == nil {
		return generateFailCallServodResponse(ctx, req.GetResource(), errors.Reason("call servod").Err())
	}
	// For labstation using port forward by ssh.
	val, err := servod.CallServod(ctx, &servod.StartServodCallRequest{
		Host:    localproxy.BuildAddr(servoHost.GetName()),
		SSHPool: c.sshPool,
		Options: &tlw.ServodOptions{
			ServodPort: servoHost.GetServodPort(),
		},
		CallMethod:    req.GetMethod(),
		CallArguments: req.GetArgs(),
		CallTimeout:   req.GetTimeout().AsDuration(),
	})
	if err != nil {
		return generateFailCallServodResponse(ctx, req.GetResource(), err)
	}
	return &tlw.CallServodResponse{
		Value: val,
		Fault: false,
	}
}

// isServoHost tells if host is servo-host.
func (c *tlwClient) isServoHost(host string) bool {
	if v, ok := c.hostTypes[host]; ok {
		return v == hostTypeServo
	}
	return false
}

// dockerClient provides docker client for target container by expected name of container.
func (c *tlwClient) dockerClient(ctx context.Context) (docker.Client, error) {
	d, err := docker.NewClient(ctx)
	return d, errors.Annotate(err, "docker client").Err()
}

// isServodContainer checks if DUT using servod-container.
// For now just simple check if servod container is provided.
// Later need distinguish when container running on the same host or remove one.
func isServodContainer(d *tlw.Dut) bool {
	return servoContainerName(d) != ""
}

// servoContainerName returns container name specified for servo-host.
func servoContainerName(d *tlw.Dut) string {
	return d.GetChromeos().GetServo().GetContainerName()
}
