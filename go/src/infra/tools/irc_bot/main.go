package main

import (
	"fmt"
	"infra/libs/gitiles"
	"infra/monitoring/looper"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luci/luci-go/client/authcli"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/cloudlogging"
	log "github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/logging/cloudlog"
	"github.com/luci/luci-go/common/logging/gologger"
	"github.com/luci/luci-go/common/logging/teelogger"
	"github.com/maruel/subcommands"
	irc "github.com/thoj/go-ircevent"
	"golang.org/x/net/context"
	"google.golang.org/cloud"
)

var (
	authOptions = auth.Options{
		Scopes: []string{
			auth.OAuthScopeEmail,
			"https://www.googleapis.com/auth/datastore",
			"https://www.googleapis.com/auth/logging.admin",
		},
	}
	commitPositionRegexp = regexp.MustCompile(
		"^cr-commit-position: refs/heads/master@{#(?P<commit_position>\\d+)\\}")
	redFlagsFormat = "\\x037%s\\x03"
)

const (
	CommitMsgRateLimit     = 5
	LoopTimeout            = time.Minute
	NumGitilesWorkers      = 10
	GoogleCloudProjectName = "chrome-infra-irc-bot"
)

// ircBot holds the IRC, gitiles, and datastore connection info,
// as well as the context containing the logger, and the current bot config.
type ircBot struct {
	g    gitilesInterface
	conn ircInterface
	ctx  context.Context
	cfg  *botConfig
	ds   dsInterface
}

func newBot(ctx context.Context, botConfig *botConfig,
	client *http.Client) *ircBot {
	g := newGitiles(botConfig.URL, NumGitilesWorkers, nil)

	dsClient, err := newDS(ctx, GoogleCloudProjectName,
		cloud.WithBaseHTTP(client))
	if err != nil {
		panic(err)
	}

	conn := makeIRC(botConfig)
	bot := &ircBot{
		g:    g,
		cfg:  botConfig,
		conn: conn,
		ctx:  ctx,
		ds:   dsClient,
	}

	conn.Connect(fmt.Sprintf("%s:%v", bot.cfg.Server, bot.cfg.Port))
	joined := make(chan bool)

	conn.AddCallback("001", func(e *irc.Event) {
		conn.Join(bot.cfg.Channel)
		joined <- true
	})
	<-joined

	return bot
}

var commitDetail = func(commit *gitiles.Commit) (string, error) {
	commitHash := commit.Commit
	if len(commitHash) < 8 {
		return "", fmt.Errorf("invalid commit hash: %s", commitHash)
	}

	email := commit.Author.Email
	subject := strings.Split(commit.Message, "\n")[0]
	body := commit.Message

	commitPosition := ""
	redFlagStrings := []string{"notry=true", "tbr="}
	redFlags := make([]string, 0)

	for _, line := range strings.Split(body, "\n") {
		for _, flaggedString := range redFlagStrings {
			if strings.HasPrefix(
				strings.ToLower(line), strings.ToLower(flaggedString)) {
				redFlags = append(redFlags, line)
			}
		}

		matches := commitPositionRegexp.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}

		commitPosition = matches[1]
	}

	if commitPosition == "" {
		commitPosition = commitHash[:8]
	}

	url := fmt.Sprintf("https://crrev.com/%s", commitPosition)

	redFlagsJoined := ""
	if len(redFlags) > 0 {
		redFlagsJoined = strings.Join(redFlags, " ")
	}

	redFlagMessage := ""
	if redFlagsJoined != "" {
		redFlagMessage = fmt.Sprintf(redFlagsFormat, redFlagsJoined)
	}

	return strings.TrimSpace(fmt.Sprintf("%s %s committed \"%s\" %s",
		url, email, subject, redFlagMessage)), nil
}

func (i *ircBot) getNewCommits(commits []*gitiles.Commit) ([]*gitiles.Commit, error) {
	lastCommit, err := i.ds.getLastCommit(i.ctx, i.cfg.ProjectName)
	if err != nil {
		return nil, err
	}

	filtered := make([]*gitiles.Commit, len(commits))
	waiter := sync.WaitGroup{}
	var numCommits int32

	for ind, commit := range commits {
		if commit.Commit == lastCommit {
			break
		}

		waiter.Add(1)
		log.Debugf(i.ctx, "Checking %s", commit.Commit)
		go func(ind int, commit *gitiles.Commit) {
			if i.shouldAnnounceCommit(commit.Commit) {
				log.Infof(i.ctx, "Found %s", commit.Commit)
				filtered[ind] = commit
				atomic.AddInt32(&numCommits, 1)
			}
			waiter.Done()
		}(ind, commit)
	}
	waiter.Wait()

	newCommits := make([]*gitiles.Commit, numCommits)
	ind := 0
	for _, commit := range filtered {
		if commit != nil {
			newCommits[ind] = commit
			ind++
		}
	}

	return newCommits, nil
}

func (i *ircBot) postCommit(detail string) {
	i.conn.Privmsg(i.cfg.Channel, detail)
}

func (i *ircBot) shouldAnnounceCommit(commit string) bool {
	files, err := i.g.getAffectedFiles(commit)
	if err != nil {
		log.Warningf(i.ctx, "Got error while getting affected files: %s", err)
		return false
	}

	for _, path := range files {
		if strings.HasPrefix(path, i.cfg.AnnouncePath) {
			log.Debugf(i.ctx, "Found path %s for commit %s",
				path, commit)
			return true
		}
	}
	return false
}

func (i *ircBot) postNewCommits() error {
	commits, err := i.g.GetLog("master")
	if err != nil {
		return err
	}

	newCommits, err := i.getNewCommits(commits)

	if err != nil {
		return err
	}

	if len(newCommits) == 0 {
		log.Debugf(i.ctx, "No new commits")
		return nil
	}

	if len(newCommits) > CommitMsgRateLimit {
		return fmt.Errorf("Way too many commits (%d).", len(newCommits))
	}

	for _, commit := range newCommits {
		detail, err := commitDetail(commit)
		if err != nil {
			log.Warningf(i.ctx, "Invalid detail for %s: %s", commit.Commit, err)
			continue
		}

		log.Debugf(i.ctx, "Posting commit %s", commit.Commit)
		i.postCommit(detail)
		log.Infof(i.ctx, "Posted message %v", detail)
	}

	return i.ds.setLastCommit(i.ctx, i.cfg.ProjectName, newCommits[0].Commit)
}

var cmdRun = &subcommands.Command{
	UsageLine: "run",
	ShortDesc: "runs the IRC bot",
	LongDesc:  "Runs the IRC bot.",
	CommandRun: func() subcommands.CommandRun {
		c := &runRun{}
		c.Flags.StringVar(&c.config, "config", "", "Config file to use")
		c.Flags.BoolVar(&c.cloudLogging, "cloudLogging", false,
			"Send logs to google cloud logging.")
		c.Flags.StringVar(&c.lastHash, "lastHash", "", "Optional initial hash.")
		return c
	},
}

type runRun struct {
	subcommands.CommandRunBase

	config       string
	cloudLogging bool
	lastHash     string
}

func (c *runRun) Run(a subcommands.Application, args []string) int {
	if c.config == "" {
		fmt.Fprintln(os.Stderr, "-config is required")
		return 1
	}

	authClient := auth.NewAuthenticator(auth.SilentLogin, authOptions)
	httpClient, err := authClient.Client()
	if err != nil {
		panic(err)
	}

	configFile, err := os.Open(c.config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}
	defer configFile.Close()

	botConfig, err := parseConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}

	ctx := gologger.Use(context.Background())

	if c.cloudLogging {
		opts := &cloudlogging.ClientOptions{
			ProjectID:    GoogleCloudProjectName,
			LogID:        botConfig.ProjectName,
			ServiceName:  "compute.googleapis.com",
			ResourceType: "service",
		}
		opts.Populate()
		cloudlogService, err := cloudlogging.NewClient(*opts, httpClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			return 1
		}

		buffered := cloudlogging.NewBuffer(
			ctx, cloudlogging.BufferOptions{}, cloudlogService)
		defer buffered.StopAndFlush()
		cloudLog := log.Get(cloudlog.Use(context.Background(), cloudlog.Config{}, buffered))
		ctx = teelogger.Use(ctx, cloudLog)
	} else {
		log.Warningf(ctx, "Not logging to cloud logging...")
	}

	cfg := &log.Config{
		Level: log.Debug,
	}
	cfg.Set(ctx)
	log.Infof(ctx, "Logging created.")

	bot := newBot(ctx, botConfig, httpClient)

	if c.lastHash != "" {
		bot.ds.setLastCommit(bot.ctx, bot.cfg.ProjectName, c.lastHash)
	}

	f := func(ctx context.Context) error {
		bot.ctx, _ = context.WithTimeout(ctx, LoopTimeout)
		defer func() {
			bot.ctx = ctx
		}()
		return bot.postNewCommits()
	}

	res := looper.Run(bot.ctx, f, time.Second*3, 3, clock.GetSystemClock())

	if res.Overruns != 0 {
		log.Warningf(bot.ctx, "Got %d overruns", res.Overruns)
	}

	if !res.Success {
		log.Errorf(bot.ctx, "Loop was not successful; errors: %d", res.Errs)
		return 1
	}
	return 0
}

var application = &subcommands.DefaultApplication{
	Name:  "irc_bot",
	Title: "Notifies IRC about new commits to a repository.",
	Commands: []*subcommands.Command{
		subcommands.CmdHelp,

		cmdRun,

		// Authentication related commands.
		authcli.SubcommandInfo(authOptions, "whoami"),
		authcli.SubcommandLogin(authOptions, "login"),
		authcli.SubcommandLogout(authOptions, "logout"),
	},
}

func main() {
	os.Exit(subcommands.Run(application, nil))
}
