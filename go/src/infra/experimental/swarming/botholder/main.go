// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary botholder manages execution of a Swarming bot inside a container.
//
// It is used to run Swarming bots on GKE for a load test. As such, it is not
// super generic and expects particular behavior from the bot.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/common/logging/sdlogger"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

var (
	prod           = flag.Bool("prod", false, "Set when running for real inside a container")
	containerImage = flag.String("container-image", "", "Name of the container image, for logs")
	oauth2Config   = flag.String("bot-oauth2-config", "/etc/swarming_config/oauth2_access_token_config.json", "Where to drop Swarming bot OAuth2 config")
	botDir         = flag.String("bot-dir", "/b/s", "Where to download and run the bot")
	swarmingHost   = flag.String("swarming-host", "chromium-swarm-dev.appspot.com", "Swarming server to fetch the bot from")
	python3        = flag.String("python3-bin", "/usr/bin/python3", "Python executable to use to run the bot")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	if *prod {
		ctx = logging.SetFactory(ctx, sdlogger.Factory(&sdlogger.Sink{Out: os.Stderr}, sdlogger.LogEntry{}, nil))
	} else {
		ctx = gologger.StdConfig.Use(ctx)
	}
	ctx = logging.SetLevel(ctx, logging.Debug)

	if err := run(ctx); err != nil {
		errors.Log(ctx, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Annotate(err, "failed to get hostname").Err()
	}
	botID := strings.Split(hostname, ".")[0]

	user, err := user.Current()
	if err != nil {
		return errors.Annotate(err, "failed to get OS user").Err()
	}

	logging.Infof(ctx, "Container image: %s", *containerImage)
	logging.Infof(ctx, "Bot ID: %s", botID)
	logging.Infof(ctx, "OS user: %s (UID=%s, GID=%s)", user.Username, user.Uid, user.Gid)

	au, err := initAuth(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to initialize auth client").Err()
	}

	botZip, err := downloadBotCode(ctx, au, botID)
	if err != nil {
		return errors.Annotate(err, "failed to download the bot code").Err()
	}

	// A server that serves OAuth2 token to the swarming bot. The bot uses them
	// to authenticate to the swarming server. Without this mechanism the bot
	// would try to use GCE Identity tokens which are not supported on GKE.
	stopAuthServer, err := launchAuthServer(ctx, au)
	if err != nil {
		return errors.Annotate(err, "failed to launch auth server").Err()
	}
	defer stopAuthServer()

	// We run as PID 1 and it is our responsibility to collect all zombie
	// processes abandoned by their parents.
	go collectZombies(ctx)

	// Shared by the signal handler and the main loop.
	var shared struct {
		sync.Mutex
		botCmd         *exec.Cmd
		sigtermPending bool
	}

	// Catch SIGTERM (sent by Kubernetes) and SIGUSR1 (sent by the shutdown.sh
	// script). The shutdown script pretends to be /sbin/shutdown. It is called by
	// Swarming bot itself when it wants to reboot the host. We'll just reboot
	// the bot itself in this case.
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, interrupts()...)
	go func() {
		for sig := range sigCh {
			logging.Warningf(ctx, "signal: Got %s", sig)
			shared.Lock()
			switch {
			case isTermSignal(sig):
				// The container is being shutdown. Tell the bot to stop, then wait for
				// it to be stopped in the main loop.
				shared.sigtermPending = true
				if shared.botCmd != nil && (shared.botCmd.Process != nil && shared.botCmd.ProcessState == nil) {
					logging.Infof(ctx, "Sending %s to the bot process", sig)
					shared.botCmd.Process.Signal(sig)
				} else {
					// The bot process is already stooped or not running yet => exit ASAP.
					logging.Infof(ctx, "No bot processing running, exiting ASAP")
					os.Exit(0)
				}
			case isUserSignal(sig):
				// The bot called /sbin/shutdown and waits for SIGTERM now. Send it.
				// Note that we will not be restarting the entire container, only the
				// bot process.
				if shared.botCmd != nil && shared.botCmd.Process != nil {
					logging.Infof(ctx, "Sending %s to the bot process", sigTerm)
					shared.botCmd.Process.Signal(sigTerm)
				}
			}
			shared.Unlock()
		}
	}()

	if !*prod {
		time.Sleep(10 * time.Minute)
		return nil
	}

	count := 0
	for {
		count += 1
		if count > 1 {
			logging.Infof(ctx, "Restarting the bot process (iteration %d)", count)
		}

		shared.Lock()
		if shared.sigtermPending {
			shared.Unlock()
			logging.Infof(ctx, "Termination signal pending, exiting")
			return nil
		}
		cmd, err := launchBot(ctx, botZip)
		shared.botCmd = cmd
		shared.Unlock()
		if err != nil {
			return errors.Annotate(err, "failed to launch the bot process").Err()
		}

		logging.Infof(ctx, "Waiting for the bot to exit...")
		if err = cmd.Wait(); err != nil {
			logging.Errorf(ctx, "The bot process exited: %s", err)
		} else {
			logging.Infof(ctx, "The bot process exited with 0 code")
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// initAuth initializes the token source used to authenticate the bot.
func initAuth(ctx context.Context) (*auth.Authenticator, error) {
	opts := chromeinfra.DefaultAuthOptions()
	if *prod {
		opts.ServiceAccountJSONPath = auth.GCEServiceAccount
		opts.SecretsDir = ""
	}
	opts.Transport = auth.NewModifyingTransport(&http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 60 * time.Second,
	}, func(req *http.Request) error {
		req.Header.Set("User-Agent", "botholder")
		return nil
	})
	au := auth.NewAuthenticator(ctx, auth.SilentLogin, opts)
	email, err := au.GetEmail()
	if err != nil {
		return nil, errors.Annotate(err, "failed to get service account email").Err()
	}
	logging.Infof(ctx, "Service account: %s", email)
	return au, nil
}

// launchAuthServer launches a server the bot uses to get its auth token.
func launchAuthServer(ctx context.Context, au *auth.Authenticator) (stop func(), err error) {
	const localAuthPort = 5555

	// See get_authentication_headers hook and oauth.oauth2_access_token_from_url.
	var cfg struct {
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	}
	cfg.URL = fmt.Sprintf("http://127.0.0.1:%d/token", localAuthPort)
	cfg.Headers = map[string]string{}

	// This file will be read by get_authentication_headers hook.
	cfgFile, err := os.Create(*oauth2Config)
	if err != nil {
		return nil, errors.Annotate(err, "error creating access token config file").Err()
	}
	defer cfgFile.Close()
	if err := json.NewEncoder(cfgFile).Encode(&cfg); err != nil {
		return nil, err
	}
	if err := cfgFile.Close(); err != nil {
		return nil, err
	}

	srv := http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", localAuthPort),
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch {
			case req.Method != "GET":
				logging.Errorf(ctx, "auth: %s %s NotAllowed", req.Method, req.URL.Path)
				http.Error(rw, "Not allowed", http.StatusMethodNotAllowed)
			case req.URL.Path != "/token":
				logging.Errorf(ctx, "auth: %s %s NotFound", req.Method, req.URL.Path)
				http.Error(rw, "Not allowed", http.StatusNotFound)
			default:
				tok, err := au.GetAccessToken(3 * time.Minute)
				if err != nil {
					logging.Errorf(ctx, "Error getting token: %s", err)
					http.Error(rw, fmt.Sprintf("error getting token: %s", err), http.StatusInternalServerError)
				} else {
					var accessTok struct {
						AccessToken string `json:"access_token"`
						ExpiresIn   int    `json:"expires_in"`
					}
					accessTok.AccessToken = tok.AccessToken
					accessTok.ExpiresIn = int(time.Until(tok.Expiry).Seconds())
					logging.Infof(ctx, "auth: sending token, expires in %d sec", accessTok.ExpiresIn)
					if err := json.NewEncoder(rw).Encode(&accessTok); err != nil {
						logging.Errorf(ctx, "Error sending token: %s", err)
					}
				}
			}
		}),
	}
	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			logging.Errorf(ctx, "Local auth server: %s", err)
		}
	}()

	return func() { srv.Close() }, nil
}

// downloadBotCode fetches the bot zip file.
func downloadBotCode(ctx context.Context, au *auth.Authenticator, botID string) (string, error) {
	logging.Infof(ctx, "Fetching bot code zip from %s to %s", *swarmingHost, *botDir)

	client, err := au.Client()
	if err != nil {
		return "", errors.Annotate(err, "getting auth client").Err()
	}

	if err := os.MkdirAll(*botDir, 0777); err != nil {
		return "", errors.Annotate(err, "failed to create bot dir").Err()
	}

	botZip := filepath.Join(*botDir, "swarming_bot.zip")
	f, err := os.Create(botZip)
	if err != nil {
		return "", errors.Annotate(err, "creating bot code zip").Err()
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://%s/bot_code", *swarmingHost), nil)
	if err != nil {
		return "", errors.Annotate(err, "failed to create GET request").Err()
	}
	req.Header.Set("X-Luci-Swarming-Bot-ID", botID)

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Annotate(err, "error sending request to get bot code").Err()
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.Reason("got HTTP %d when fetching bot code zip", resp.StatusCode).Err()
	}

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", errors.Annotate(err, "failed to fetch bot code zip").Err()
	}
	if err := f.Close(); err != nil {
		return "", errors.Annotate(err, "failed to flush bot code zip").Err()
	}

	logging.Infof(ctx, "Fetched bot code zip, %d bytes: %s", n, botZip)
	return botZip, nil
}

// launchBot launches the bot process.
func launchBot(ctx context.Context, botZip string) (*exec.Cmd, error) {
	logging.Infof(ctx, "Launching bot %s %s", *python3, botZip)
	cmd := exec.CommandContext(ctx, *python3, botZip)
	cmd.Dir = filepath.Dir(botZip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	// Disable attempts of Swarming bot to register itself in auto-start.
	cmd.Env = append(cmd.Env, "SWARMING_EXTERNAL_BOT_SETUP=1")

	// Shut up noisy shivas error logs in bot_config.py.
	cmd.Env = append(cmd.Env, "SHIVAS_BINARY=-")
	cmd.Env = append(cmd.Env, "SHIVAS_RO_CRED=-")

	// Make CIPD do less stuff in parallel, to eat less RAM.
	cmd.Env = append(cmd.Env, "CIPD_MAX_THREADS=1")
	cmd.Env = append(cmd.Env, "CIPD_PARALLEL_DOWNLOADS=0")

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	logging.Infof(ctx, "Bot is running as PID %d", cmd.Process.Pid)
	return cmd, nil
}
