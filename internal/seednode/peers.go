package seednode

import (
	"github.com/tendermint/tendermint/p2p"
	"github.com/terran-stakers/seednode-test/internal/tendermint"
)

func DialPeers(peers []string, sw *tendermint.Switch, reactor *tendermint.SeedNodeReactor) {
	netAddresses, errors := p2p.NewNetAddressStrings(peers)
	if len(errors) > 0 {
		logger.Error("Invalid peers", "error", errors)
		return
	}

	err := tendermint.DialPeers(netAddresses, sw.NetAddress(), reactor)
	if err != nil {
		panic(err)
	}
}
