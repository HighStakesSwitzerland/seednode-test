package main

import (
	"github.com/tendermint/tendermint/libs/log"
	"github.com/terran-stakers/seednode-test/internal/http"
	"github.com/terran-stakers/seednode-test/internal/seednode"
  "github.com/terran-stakers/seednode-test/internal/tendermint"
  "os"
)

var (
	logger  = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	sw      *tendermint.Switch
	reactor *tendermint.SeedNodeReactor
)

func main() {
	seedConfig, nodeKey := seednode.InitConfig()

	logger.Info("Starting Seed Node...")
	sw, reactor = seednode.StartSeedNode(seedConfig, nodeKey)

	logger.Info("Starting Web Server...")
	http.StartWebServer(seedConfig, sw, reactor)

	sw.Wait() // block
}
