package seednode

import (
	"fmt"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
	"github.com/tendermint/tendermint/version"
	"github.com/terran-stakers/seednode-test/internal/tendermint"
	"os"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "seednode")
)

func StartSeedNode(seedConfig TSConfig, nodeKey p2p.NodeKey) (*tendermint.Switch, *tendermint.SeedNodeReactor) {
	cfg := config.DefaultP2PConfig()
	cfg.AllowDuplicateIP = true

	protocolVersion :=
		p2p.NewProtocolVersion(
			version.P2PProtocol,
			version.BlockProtocol,
			0,
		)

	// NodeInfo gets info on your node
	nodeInfo := p2p.DefaultNodeInfo{
		ProtocolVersion: protocolVersion,
		DefaultNodeID:   nodeKey.ID(),
		ListenAddr:      seedConfig.ListenAddress,
		Network:         seedConfig.ChainId,
		Version:         "1.0.0",
		Channels:        []byte{pex.PexChannel},
		Moniker:         fmt.Sprintf("%s-seed", seedConfig.ChainId),
	}

	addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeInfo.DefaultNodeID, nodeInfo.ListenAddr))
	if err != nil {
		panic(err)
	}
	connConfig := tendermint.NewMConnConfig(cfg)
	transport := tendermint.NewMultiplexTransport(nodeInfo, nodeKey, connConfig)
	if err := transport.Listen(*addr); err != nil {
		panic(err)
	}

	pexReactor := tendermint.NewReactor()
	sw := tendermint.NewSwitch(cfg, transport)
	sw.SetNodeKey(&nodeKey)
	sw.AddReactor("pex", pexReactor)
	sw.SetLogger(logger.With("module", "switch"))
	pexReactor.SetLogger(logger.With("module", "pex"))

	// last
	sw.SetNodeInfo(nodeInfo)

	tmos.TrapSignal(logger, func() {
		logger.Info("shutting down...")
		err := sw.Stop()
		if err != nil {
			panic(err)
		}
	})

	err = sw.Start()
	if err != nil {
		panic(err)
	}

	return sw, pexReactor
}
