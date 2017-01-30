package main

import (
	"golang.org/x/net/context"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
)

var (
	newDS = func(ctx context.Context, name string, opts cloud.ClientOption) (dsInterface, error) {
		cl, err := datastore.NewClient(ctx, name, opts)
		return &dsImpl{cl}, err
	}
)

type dsInterface interface {
	getLastCommit(context.Context, string) (string, error)
	setLastCommit(context.Context, string, string) error
}

type dsImpl struct {
	*datastore.Client
}

type lastCommit struct {
	LastHash string
}

func (d *dsImpl) getLastCommit(ctx context.Context, projectName string) (string, error) {
	commit := lastCommit{}
	err := d.Get(ctx, datastore.NewKey(
		ctx, "lastCommit", projectName, 0, nil), &commit)

	if err != nil {
		return "", err
	}

	return commit.LastHash, nil
}

func (d *dsImpl) setLastCommit(ctx context.Context, projectName, hash string) error {
	commit := lastCommit{
		LastHash: hash,
	}
	_, err := d.Put(ctx, datastore.NewKey(
		ctx, "lastCommit", projectName, 0, nil), &commit)

	if err != nil {
		return err
	}
	return nil
}
