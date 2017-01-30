package main

import (
	"encoding/json"
	"io"
)

type botConfig struct {
	URL          string `json:"gitiles_url"`
	Nickname     string `json:"nickname"`
	AnnouncePath string `json:"announce_path"`
	ProjectName  string `json:"project_name"`
	Channel      string `json:"channel"`
	Server       string `json:"server"`
	Port         int    `json:"port"`
	UseSSL       bool   `json:"use_ssl"`
}

func parseConfig(file io.Reader) (*botConfig, error) {
	result := botConfig{}
	err := json.NewDecoder(file).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
