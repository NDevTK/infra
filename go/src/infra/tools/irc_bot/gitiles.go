package main

import (
	"net/http"

	"github.com/luci/luci-go/common/stringset"
	"infra/libs/gitiles"
)

var (
	newGitiles = func(url string, maxConnections int, client *http.Client) gitilesInterface {
		return &gitilesImpl{gitiles.NewGitiles(url, maxConnections, client)}
	}
)

type gitilesInterface interface {
	getAffectedFiles(string) ([]string, error)
	GetLog(string) ([]*gitiles.Commit, error)
}

type gitilesImpl struct {
	*gitiles.Gitiles
}

func (g *gitilesImpl) getAffectedFiles(commitHash string) ([]string, error) {
	commit, err := g.GetCommit(commitHash)
	if err != nil {
		return nil, err
	}

	set := stringset.New(len(commit.TreeDiff))
	for _, diff := range commit.TreeDiff {
		set.Add(diff.OldPath)
		set.Add(diff.NewPath)
	}

	set.Del("/dev/null")
	return set.ToSlice(), nil
}
