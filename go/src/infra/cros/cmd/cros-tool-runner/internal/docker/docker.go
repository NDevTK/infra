// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package docker provide abstaraction to pull/start/stop/remove docker image.
// Package uses docker-cli from running host.
package docker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros-tool-runner/internal/common"
)

const (
	// Default fallback docket tag.
	DefaultImageTag = "stable"
)

// Docker holds data to perform the docker manipulations.
type Docker struct {
	// Requested docker image, if not exist then use FallbackImageName.
	RequestedImageName string
	// Registry to auth for docker interactions.
	Registry string
	// token to token
	Token string
	// Fall back docker image name. Used if RequestedImageName is empty or image not found.
	FallbackImageName string
	// ServicePort tells which port need to bing bind from docker to the host.
	// Bind is always to the first free port.
	ServicePort int
	// Run container in detach mode.
	Detach bool
	// Name to be assigned to the container - should be unique.
	Name string
	// ExecCommand tells if we need run special command when we start container.
	ExecCommand []string
	// Attach volumes to the docker image.
	Volumes []string
	// PortMappings is a list of "host port:docker port" or "docker port" to publish.
	PortMappings []string

	// Successful pulled docker image.
	pulledImage string
	// Started container ID.
	containerID string
	// Network used for running container.
	Network string
}

// HostPort returns the port which the given docker port maps to.
func (d *Docker) MatchingHostPort(ctx context.Context, dockerPort string) (string, error) {
	cmd := exec.Command("docker", "port", d.Name, dockerPort)
	stdout, stderr, err := common.RunWithTimeout(ctx, cmd, 2*time.Minute, true)
	if err != nil {
		log.Printf(fmt.Sprintf("Could not find port %v for %v: %v", dockerPort, d.Name, err), stdout, stderr)
		return "", errors.Annotate(err, "find mapped port").Err()
	}

	// Expected stdout is of the form "0.0.0.0:12345\n".
	port := strings.TrimPrefix(stdout, "0.0.0.0:")
	port = strings.TrimSuffix(port, "\n")
	return port, nil
}

// PullImage pulls docker image.
// The first we try to pull with required tag and if fail then use default tag. Image with default tag always present in repo.
func (d *Docker) PullImage(ctx context.Context) (err error) {
	if d.RequestedImageName != "" {
		d.pulledImage = d.RequestedImageName
		if err = pullImage(ctx, d.pulledImage); err == nil {
			return nil
		}
	}
	if d.FallbackImageName != "" {
		d.pulledImage = d.FallbackImageName
		if err = pullImage(ctx, d.pulledImage); err == nil {
			return nil
		}
	}
	if err != nil {
		return errors.Annotate(err, "pull image").Err()
	}
	return errors.Reason("pull image: failed").Err()
}

// pullImage pulls image by docker-cli.
// docker-cli has to have permission to download images from required repos.
func pullImage(ctx context.Context, image string) error {
	cmd := exec.Command("docker", "pull", image)
	stdout, stderr, err := common.RunWithTimeout(ctx, cmd, 2*time.Minute, true)
	common.PrintToLog(fmt.Sprintf("Pull image %q", image), stdout, stderr)
	if err != nil {
		log.Printf("pull image %q: failed with error: %s", image, err)
		return errors.Annotate(err, "Pull image").Err()
	}
	log.Printf("pull image %q: successful pulled.", image)
	return nil
}

// Auth with docker registry so that pulling and stuff works.
func (d *Docker) Auth(ctx context.Context) (err error) {
	if d.Token == "" {
		log.Printf("no token was provided so skipping docker auth.")
		return nil
	}
	if d.Registry == "" {
		return errors.Reason("docker auth: failed").Err()
	}

	if err = auth(ctx, d.Registry, d.Token); err != nil {
		return errors.Annotate(err, "docker auth").Err()
	}
	return nil
}

// auth authorizes the current process to the given registry, using keys on the drone.
// This will give permissions for pullImage to work :)
func auth(ctx context.Context, registry string, token string) error {
	cmd := exec.Command("docker", "login", "-u", "oauth2accesstoken",
		"-p", token, registry)
	stdout, stderr, err := common.RunWithTimeout(ctx, cmd, 1*time.Minute, true)
	common.PrintToLog("Login", stdout, stderr)
	if err != nil {
		return errors.Annotate(err, "failed running 'docker login'").Err()
	}
	log.Printf("login successful!")
	return nil
}

// Remove removes the containers with matched name.
func (d *Docker) Remove(ctx context.Context) error {
	if d == nil {
		return nil
	}
	// Use force to avoid any un-related issues.
	cmd := exec.Command("docker", "rm", "--force", d.Name)
	stdout, stderr, err := common.RunWithTimeout(ctx, cmd, time.Minute, true)
	common.PrintToLog(fmt.Sprintf("Remove container %q", d.Name), stdout, stderr)
	if err != nil {
		log.Printf("remove container %q failed with error: %s", d.Name, err)
		return errors.Annotate(err, "remove container %q", d.Name).Err()
	}
	log.Printf("remove container %q: done.", d.Name)
	return nil
}

// Run docker image.
// The step will create container and start server inside or execution CLI.
func (d *Docker) Run(ctx context.Context, block bool) error {
	out, err := d.runDockerImage(ctx, block)
	if err != nil {
		return errors.Annotate(err, "run docker %q", d.Name).Err()
	}
	if d.Detach {
		d.containerID = strings.TrimSuffix(out, "\n")
		log.Printf("Run docker %q: container Id: %q.", d.Name, d.containerID)
	}
	return nil
}

func (d *Docker) runDockerImage(ctx context.Context, block bool) (string, error) {
	args := []string{"run"}
	if d.Detach {
		args = append(args, "-d")
	}
	args = append(args, "--name", d.Name)
	for _, v := range d.Volumes {
		args = append(args, "-v")
		args = append(args, v)
	}
	// Set to automatically remove the container when it exits.
	args = append(args, "--rm")
	if d.Network != "" {
		args = append(args, "--network", d.Network)
	}

	// Publish in-docker ports; any without an explicit mapping will need to be looked up later.
	if len(d.PortMappings) != 0 {
		args = append(args, "-p")
		args = append(args, d.PortMappings...)
	}

	args = append(args, d.pulledImage)
	if len(d.ExecCommand) > 0 {
		args = append(args, d.ExecCommand...)
	}

	cmd := exec.Command("docker", args...)
	so, se, err := common.RunWithTimeout(ctx, cmd, time.Hour, block)
	common.PrintToLog(fmt.Sprintf("Run docker image %q", d.Name), so, se)
	return so, errors.Annotate(err, "run docker image %q: %s", d.Name, se).Err()
}

// CreateImageName creates docker image name from repo-path and tag.
func CreateImageName(repoPath, tag string) string {
	return fmt.Sprintf("%s:%s", repoPath, tag)
}

// CreateImageNameFromInputInfo creates docker image name from input info.
//
// If info is empty then return empty name.
// If one of the fields empty then use related default value.
func CreateImageNameFromInputInfo(di *api.DutInput_DockerImage, defaultRepoPath, defaultTag string) string {
	if di == nil {
		return ""
	}
	if di.GetRepositoryPath() == "" && di.GetTag() == "" {
		return ""
	}
	repoPath := di.GetRepositoryPath()
	if repoPath == "" {
		repoPath = defaultRepoPath
	}
	tag := di.GetTag()
	if tag == "" {
		tag = defaultTag
	}
	if repoPath == "" || tag == "" {
		panic("Default repository path or tag for docker image was not passed.")
	}
	return CreateImageName(repoPath, tag)
}
