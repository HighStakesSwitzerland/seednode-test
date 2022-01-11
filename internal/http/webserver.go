package http

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/terran-stakers/seednode-test/internal/seednode"
	"github.com/terran-stakers/seednode-test/internal/tendermint"
	"net/http"
	"os"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "web")
)

type WebResources struct {
	Res   embed.FS
	Files map[string]string
}

func StartWebServer(seedConfig seednode.TSConfig, sw *tendermint.Switch, reactor *tendermint.SeedNodeReactor) {
	// serve endpoint
	http.HandleFunc("/api/peers", func(w http.ResponseWriter, r *http.Request) {
		handleOperation(w, r, sw, reactor)
	})

	// start web server in non-blocking
	go func() {
		err := http.ListenAndServe(":"+seedConfig.HttpPort, nil)
		logger.Info("HTTP Server started", "port", seedConfig.HttpPort)
		if err != nil {
			panic(err)
		}
	}()
}

func handleOperation(w http.ResponseWriter, r *http.Request, sw *tendermint.Switch, reactor *tendermint.SeedNodeReactor) {
	if r.Method == "GET" {
		writePeers(w)
	} else if r.Method == "POST" {
		handlePeers(w, r, sw, reactor)
	}
}

func handlePeers(w http.ResponseWriter, r *http.Request, sw *tendermint.Switch, reactor *tendermint.SeedNodeReactor) {

	var peers []string
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&peers)

	if err != nil {
		_, _ = w.Write([]byte("Bad Request " + err.Error()))
		w.WriteHeader(400)
		return
	}

	logger.Info(fmt.Sprintf("Received %d peers to dial", len(peers)))

	// emit peers
	seednode.DialPeers(peers, sw, reactor)

	w.WriteHeader(200)
}

func writePeers(w http.ResponseWriter) {
	marshal, err := json.Marshal("")
	if err != nil {
		logger.Info("Failed to marshal peers list")
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		return
	}
}
