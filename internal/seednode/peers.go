package seednode

import (
	"fmt"
	"github.com/tendermint/tendermint/p2p"
)

var (
	peerList []*Peer
)

type Peer struct {
	NodeId  p2p.ID `json:"nodeId"`
	Moniker string `json:"moniker"`
}

type MyPeer struct {
	p2p.Peer
	sw p2p.Switch
}

func DialPeers(peers []string, sw *p2p.Switch, reactor *SeedNodeReactor) {
	netAddresses, errors := p2p.NewNetAddressStrings(peers)
	if len(errors) > 0 {
		panic(errors)
	}

	p := &MyPeer{nil, *sw}
	err := DialPeer(netAddresses, p, reactor)
	if err != nil {
		panic(err)
	}
}

/*
	Returns the current reactor peers. As in seed mode the pex module disconnects quickly, this list can grow and shrink
	according to the current connexions
*/
func GetPeers(peers []p2p.Peer) []*Peer {
	if len(peers) > 0 {
		logger.Info(fmt.Sprintf("Address book contains %d new peers", len(peers)), "peers", peers)
		peerList = p2pPeersToPeerList(peers)
		return peerList
	}
	return nil
}

func p2pPeersToPeerList(list []p2p.Peer) []*Peer {
	var _peers []*Peer
	for _, p := range list {
		_peers = append(_peers, &Peer{
			NodeId:  p.NodeInfo().ID(),
			Moniker: p.NodeInfo().(p2p.DefaultNodeInfo).Moniker,
		})
	}
	return _peers
}

func (p MyPeer) ID() p2p.ID {
	return p.sw.NodeInfo().ID()
}
func (p MyPeer) NodeInfo() p2p.NodeInfo {
	return p.sw.NodeInfo()
}
