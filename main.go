package main

import (
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/terran-stakers/seednode-test/internal/http"
	"github.com/terran-stakers/seednode-test/internal/seednode"
	"os"
)

var (
	logger  = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	sw      *p2p.Switch
	reactor *seednode.SeedNodeReactor
)

func main() {
	seedConfig, nodeKey := seednode.InitConfig()

	logger.Info("Starting Seed Node...")
	sw, reactor = seednode.StartSeedNode(seedConfig, nodeKey)

	logger.Info("Starting Web Server...")
	http.StartWebServer(seedConfig, sw, reactor)

	sw.Wait() // block
}
