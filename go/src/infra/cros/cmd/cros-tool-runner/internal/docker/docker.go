// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package docker provide abstaraction to pull/start/stop/remove docker image.
// Package uses docker-cli from running host.
package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/mitchellh/go-homedir"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros-tool-runner/internal/common"
)

const (
	// Default fallback docket tag.
	DefaultImageTag  = "stable"
	basePodmanConfig = "/run/containers/0/auth.json"
	baseDockerConfig = "~/.docker/config.json"
	dockerRegistry   = "us-docker.pkg.dev"
	lockFile         = "/var/lock/go-lock.lock"
)

// Docker holds data to perform the docker manipulations.
type Docker struct {
	// Requested docker image, if not exist then use FallbackImageName.
	RequestedImageName string
	// Registry to auth for docker interactions.
	Registry string
	// token to token
	TokenFile string
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

// Auth with docker registry so that pulling and stuff works.
func (d *Docker) Auth(ctx context.Context) (err error) {
	if d.TokenFile == "" {
		log.Printf("no token was provided so skipping docker auth.")
		return nil
	}
	if d.Registry == "" {
		return errors.Reason("docker auth: failed").Err()
	}

	token, err := GCloudToken(ctx, d.TokenFile, false)
	if err = auth(ctx, d.Registry, token); err != nil {
		// If the login fails, force a full token regen.
		token, err := GCloudToken(ctx, d.TokenFile, true)
		if err != nil {
			return errors.Annotate(err, "GCloudToken force").Err()
		}
		// Then try to auth again, and if THAT fails, err time.
		if err = auth(ctx, d.Registry, token); err != nil {
			return errors.Annotate(err, "docker auth").Err()
		}
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
func (d *Docker) Run(ctx context.Context, block bool, netbind bool) error {
	out, err := d.runDockerImage(ctx, block, netbind)
	if err != nil {
		return errors.Annotate(err, "run docker %q", d.Name).Err()
	}
	if d.Detach {
		d.containerID = strings.TrimSuffix(out, "\n")
		log.Printf("Run docker %q: container Id: %q.", d.Name, d.containerID)
	}
	return nil
}

func (d *Docker) runDockerImage(ctx context.Context, block bool, netbind bool) (string, error) {
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

	// give access to net_raw so things like `ping` can work in the container.
	if netbind == true {
		args = append(args, "--cap-add=NET_RAW")
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

func maybeFindToken(forceNewAuth bool) (string, error) {
	err, authFileDir := authFile(forceNewAuth)
	if err == nil && authFileDir != "" {
		log.Println("Previously authenticated authorization token found. Skipping auth.")
		return readToken(authFileDir)
	}
	return "", err
}

// GCloudToken will try to return the gcloud token for `docker login`.
func GCloudToken(ctx context.Context, keyfile string, forceNewAuth bool) (string, error) {
	// This method will first try to get an existing login token from the known token files.
	// If it does not exist it will gcloud auth, then get the token.
	// the `gcloud auth` commands will be a system level lock command to avoid DB races (which caused crashes).
	// Thus other CTR instances will be held in line until the one with the lock finishes.
	// Only the first execution on the drone (or after a 24hr expiration time) should ever need to `auth`.
	if token, err := maybeFindToken(forceNewAuth); token != "" {
		return token, err
	}

	log.Println("Attempting to gcloud auth.")
	// Get the lock, which is a blocking call to wait for the lock.
	fileLock := flock.New(lockFile)
	err := fileLock.Lock()
	if err != nil {
		return "", errors.Annotate(err, "failed to get FLock prior to gcloud calls").Err()
	}
	defer fileLock.Unlock()
	log.Println("FLock obtained")

	// Check the Auth again. Its possible someone else was authing as we waited for the lock.
	if token, err := maybeFindToken(forceNewAuth); token != "" {
		return token, err
	}
	// Finally, if nothing was there, and we have the lock, auth/return the str.
	return gcloudAuth(ctx, keyfile)
}

// gcloudAuth will run the `gcloud auth` cmd and return the access-token.
func gcloudAuth(ctx context.Context, keyfile string) (string, error) {
	err := activateAccount(ctx, keyfile)
	if err != nil {
		return "", fmt.Errorf("could not activate account: %s", err)
	}

	cmd := exec.Command("gcloud", "auth", "print-access-token")
	out, _, err := common.RunWithTimeout(ctx, cmd, 5*time.Minute, true)
	if err != nil {
		return "", errors.Annotate(err, "failed getting gcloud access token.").Err()
	}
	return out, nil
}

// authFile returns the gcloud auth file if found, else ""
func authFile(forceNewAuth bool) (error, string) {
	if forceNewAuth {
		return nil, ""
	}
	dockerConfigPath, _ := homedir.Expand(baseDockerConfig)
	podmanConfigPath, _ := homedir.Expand(basePodmanConfig)

	for _, dir := range []string{podmanConfigPath, dockerConfigPath} {
		log.Printf("Checking for authfile: %s\n", dir)
		if f, err := os.Stat(dir); err == nil {
			modifiedTime := f.ModTime()
			if time.Now().Sub(modifiedTime).Hours() >= 24 {
				log.Println("Auth Token is more than 24 hours old, forcing a refresh.")
				return nil, ""
			}
			log.Println("Found Auth file.")
			return nil, dir
		} else if errors.Is(err, os.ErrNotExist) {
			continue
		} else {
			return err, ""
		}
	}
	return nil, ""
}

// readToken will read the given json, and return the decoded oath token for docker login.
func readToken(dir string) (string, error) {
	log.Println("Reading docker login oath token from the found config file.")
	jsonFile, err := os.Open(dir)
	if err != nil {
		log.Printf("Error reading tokeon json file: %s", err)
		return "", err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	// ugly parse the json.
	f := result["auths"].(map[string]interface{})[dockerRegistry].(map[string]interface{})["auth"]
	str := fmt.Sprintf("%v", f)

	// convert magic to the usable str for dockerLogin.
	decode, _ := base64.StdEncoding.DecodeString(str)
	s := string(decode)
	s = strings.ReplaceAll(s, "oauth2accesstoken:", "")

	return s, nil
}

// activateAccount actives the gcloud service account using the given keyfile
func activateAccount(ctx context.Context, keyfile string) error {
	log.Println("Obtaining oath token from gcloud auth.")
	if _, err := os.Stat(keyfile); err == nil {
		// keyfile exists
		cmd := exec.Command("gcloud", "auth", "activate-service-account",
			fmt.Sprintf("--key-file=%v", keyfile))
		out, stderr, err := common.RunWithTimeout(ctx, cmd, 5*time.Minute, true)
		if err != nil {
			log.Printf("Failed running gcloud auth: %s\n%s", err, stderr)
			return errors.Annotate(err, "gcloud auth").Err()
		}
		log.Printf("gcloud auth completed. Result: %s", out)
	} else if os.IsNotExist(err) {
		// keyfile doesn't exist.
		// For this case, we will assume that env has account with proper permissions.
		log.Printf("Skipping gcloud auth as keyfile does not exist")
	} else {
		// keyfile may or may not exist. See err for details.
		return errors.Annotate(err, "error with keyfile").Err()
	}
	return nil

}
