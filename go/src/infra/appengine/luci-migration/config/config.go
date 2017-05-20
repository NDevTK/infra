package config

import (
	"golang.org/x/net/context"

	"github.com/luci/luci-go/luci_config/server/cfgclient"
	"github.com/luci/luci-go/luci_config/server/cfgclient/textproto"
)

// UNSET is a value of an enum field that was not set.
const UNSET = 0

// Get returns currently imported config.
func Get(c context.Context) (*Config, error) {
	var cfg Config
	return &cfg, cfgclient.Get(
		c,
		cfgclient.AsService,
		cfgclient.CurrentServiceConfigSet(c),
		"config.cfg",
		textproto.Message(&cfg),
		nil,
	)
}
