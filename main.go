package main

import (
	"github.com/highstakesswitzerland/seednode-test/internal/http"
	"github.com/highstakesswitzerland/seednode-test/internal/seednode"
	"github.com/highstakesswitzerland/seednode-test/internal/tendermint"
	"github.com/tendermint/tendermint/libs/log"
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
