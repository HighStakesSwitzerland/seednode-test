package http

import (
	"github.com/tendermint/tendermint/p2p"
)

type NgPeer struct {
	NodeInfo  p2p.NodeInfo `mapstructure:"NodeInfo"`
	LastValue int          `mapstructure:"last_value"`
	Status    string       `mapstructure:"status"`
	Error     string       `mapstructure:"error"`
}
