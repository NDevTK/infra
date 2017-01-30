package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"infra/libs/gitiles"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	irc "github.com/thoj/go-ircevent"
	"golang.org/x/net/context"
)

type testData struct {
	Name   string   `json:"name"`
	Lines  []string `json:"lines"`
	Result []string `json:"result"`
}

type suiteData struct {
	BaseData struct {
		Commit        string `json:"commit"`
		Email         string `json:"email"`
		MessagePrefix string `json:"message_prefix"`
	} `json:"base_data"`
	Tests []testData `json:"tests"`
}

func TestFormat(t *testing.T) {
	suiteData := suiteData{}
	file, err := os.Open("test_data.json")
	if err != nil {
		panic(err)
	}

	err = json.NewDecoder(file).Decode(&suiteData)
	if err != nil {
		panic(err)
	}

	oldFormat := redFlagsFormat
	redFlagsFormat = "BAD %s BAD"
	defer func() { redFlagsFormat = oldFormat }()

	Convey("Format Tests", t, func() {
		Convey("General tests", func() {
			testFunc := func(data testData) {
				res, err := commitDetail(&gitiles.Commit{
					Commit: suiteData.BaseData.Commit,
					Author: gitiles.CommitUser{
						Email: suiteData.BaseData.Email,
					},
					Message: suiteData.BaseData.MessagePrefix + "\n" + strings.Join(data.Lines, "\n"),
				})
				So(err, ShouldBeNil)
				So(res, ShouldEqual, strings.Join(data.Result, " "))
			}

			for _, test := range suiteData.Tests {
				Convey(test.Name, func() {
					testFunc(test)
				})
			}
		})

		Convey("Error tests", func() {
			_, err := commitDetail(&gitiles.Commit{
				Commit: "a",
			})
			So(err, ShouldNotBeNil)
		})
	})
}

type dsMockGC struct {
	getErr error
	commit *gitiles.Commit
}

func (d *dsMockGC) getLastCommit(context.Context, string) (string, error) {
	return d.commit.Commit, d.getErr
}

func (d *dsMockGC) setLastCommit(context.Context, string, string) error {
	panic("should not be called")
}

type gMockGC struct {
	commits []*gitiles.Commit
	err     error
	files   map[string][]string
}

func (g *gMockGC) getAffectedFiles(commitHash string) ([]string, error) {
	if g.err != nil {
		return nil, g.err
	}

	if files, ok := g.files[commitHash]; ok {
		return files, nil
	}

	return nil, nil
}

func (g *gMockGC) GetLog(string) ([]*gitiles.Commit, error) { panic("should not be called") }

func TestGetCommits(t *testing.T) {
	t.Parallel()

	Convey("Get New Commits", t, func() {
		commits := []*gitiles.Commit{
			{Commit: "a"},
			{Commit: "b"},
			{Commit: "c"},
		}
		files := map[string][]string{
			"a": {"foo/bar/baz", "foo/baz/bam"},
			"b": {"bam/bar"},
			"c": {"fizz"},
		}

		g := &gMockGC{commits: commits, err: nil, files: files}
		d := &dsMockGC{nil, commits[len(commits)-1]}
		cfg := &botConfig{AnnouncePath: "foo"}

		i := &ircBot{
			ctx: context.Background(),
			cfg: cfg,
			g:   g,
			ds:  d,
		}

		Convey("basic test", func() {
			res, err := i.getNewCommits(commits)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, commits[:1])
		})
		Convey("path does not match", func() {
			cfg.AnnouncePath = "BLERG"
			res, err := i.getNewCommits(commits)
			So(err, ShouldBeNil)
			So(res, ShouldBeEmpty)
		})
		Convey("middle item", func() {
			cfg.AnnouncePath = "bam"
			res, err := i.getNewCommits(commits)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, commits[1:2])
		})
		Convey("partial path match", func() {
			cfg.AnnouncePath = "foo/bar"
			res, err := i.getNewCommits(commits)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, commits[:1])
		})
		Convey("ds get error", func() {
			d.getErr = errors.New("BAD THINGS")
			res, err := i.getNewCommits(commits)
			So(err, ShouldNotBeNil)
			So(res, ShouldBeNil)
		})
	})
}

type dsMockPNC struct {
	getErr    error
	setErr    error
	commit    *gitiles.Commit
	setValues map[string][]string
}

func (d *dsMockPNC) getLastCommit(context.Context, string) (string, error) {
	return d.commit.Commit, d.getErr
}

func (d *dsMockPNC) setLastCommit(ctx context.Context, name string, value string) error {
	if d.setErr != nil {
		return d.setErr
	}

	v, ok := d.setValues[name]
	if !ok {
		v = []string{value}
	} else {
		v = append(v, value)
	}

	d.setValues[name] = v
	return nil
}

type gMockPNC struct {
	logErr  error
	commits *[]*gitiles.Commit
	err     error
	files   *map[string][]string
}

func (g *gMockPNC) getAffectedFiles(commitHash string) ([]string, error) {
	if g.err != nil {
		return nil, g.err
	}

	if files, ok := (*g.files)[commitHash]; ok {
		return files, nil
	}

	return nil, nil
}

func (g *gMockPNC) GetLog(string) ([]*gitiles.Commit, error) {
	return *g.commits, g.logErr
}

type iMockPNC struct {
	sentMessages map[string][]string
}

func (i *iMockPNC) Connect(string) error                          { panic("should not be called") }
func (i *iMockPNC) AddCallback(string, func(e *irc.Event)) string { panic("should not be called") }
func (i *iMockPNC) Join(string)                                   { panic("should not be called") }
func (i *iMockPNC) Privmsg(channel, msg string) {
	v, ok := i.sentMessages[channel]
	if !ok {
		v = []string{msg}
	} else {
		v = append(v, msg)
	}

	i.sentMessages[channel] = v
}

func TestPostNewCommits(t *testing.T) {
	t.Parallel()

	Convey("PostNewCommits", t, func() {
		commits := []*gitiles.Commit{
			{Commit: "a"},
			{Commit: "b"},
			{Commit: "c"},
		}
		files := map[string][]string{
			"a": {"foo/bar/baz", "foo/baz/bam"},
			"b": {"bam/bar"},
			"c": {"fizz"},
		}

		g := &gMockPNC{commits: &commits, err: nil, files: &files}
		d := &dsMockPNC{
			getErr:    nil,
			setErr:    nil,
			commit:    commits[len(commits)-1],
			setValues: make(map[string][]string),
		}
		irc := &iMockPNC{make(map[string][]string)}
		cfg := &botConfig{
			AnnouncePath: "foo",
			Channel:      "#hashtag",
			ProjectName:  "chrome_infra_secret_project",
		}

		i := &ircBot{
			ctx:  context.Background(),
			cfg:  cfg,
			g:    g,
			ds:   d,
			conn: irc,
		}

		oldDetail := commitDetail
		defer func() { commitDetail = oldDetail }()
		detailErr := (error)(nil)
		commitDetail = func(commit *gitiles.Commit) (string, error) {
			return fmt.Sprintf("TEST MESSAGE: %s", commit.Commit), detailErr
		}

		Convey("basic", func() {
			err := i.postNewCommits()

			So(err, ShouldBeNil)
			So(len(irc.sentMessages), ShouldEqual, 1)
			So(irc.sentMessages["#hashtag"], ShouldResemble, []string{"TEST MESSAGE: a"})
			So(len(d.setValues), ShouldEqual, 1)
			So(d.setValues["chrome_infra_secret_project"], ShouldResemble, []string{"a"})
		})
		Convey("too many commits", func() {
			lastCommit := commits[len(commits)-1]
			commits = commits[:len(commits)-1]

			for i := 0; i < CommitMsgRateLimit; i++ {
				hash := fmt.Sprintf("comm:%s", i)
				commits = append(commits, &gitiles.Commit{Commit: hash})
				files[hash] = files["a"][:]
			}
			commits = append(commits, lastCommit)

			err := i.postNewCommits()
			So(err, ShouldNotBeNil)
			So(irc.sentMessages, ShouldBeEmpty)
			So(d.setValues, ShouldBeEmpty)
		})
		Convey("no new commits", func() {
			commits = commits[:0]

			err := i.postNewCommits()
			So(err, ShouldBeNil)
			So(irc.sentMessages, ShouldBeEmpty)
			So(d.setValues, ShouldBeEmpty)
		})
	})
}
