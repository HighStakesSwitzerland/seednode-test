package http

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/highstakesswitzerland/seednode-test/internal/seednode"
	"github.com/highstakesswitzerland/seednode-test/internal/tendermint"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
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
		getAllPeers(w, sw)
	} else if r.Method == "POST" {
		handlePeers(w, r, reactor)
	} else if r.Method == "DELETE" {
		deletePeers(r, reactor)
	} else {
		w.WriteHeader(500)
	}
}

func deletePeers(r *http.Request, reactor *tendermint.SeedNodeReactor) {
	id := r.URL.Query().Get("id")
	if len(id) > 0 {
		if p := reactor.Switch.GetPersistentPeers().Get(p2p.ID(id)); p != nil {
			reactor.Switch.GetPersistentPeers().Remove(p)
		}
	} else {
		reactor.Switch.GetPersistentPeers().Clear()
	}

}

func handlePeers(w http.ResponseWriter, r *http.Request, reactor *tendermint.SeedNodeReactor) {
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
	errors := tendermint.DialPeers(peers, reactor)

	if errors != nil && len(errors) > 0 {
		writeErrorToResponse(w, errors)
	}
}

func getAllPeers(w http.ResponseWriter, sw *tendermint.Switch) {
	writePeersToResponse(w, peersToNgPeers(sw.GetPersistentPeers().List()))
}

func writePeersToResponse(w http.ResponseWriter, peers []NgPeer) {
	if len(peers) == 0 {
		_, _ = w.Write([]byte("[]"))
		return
	}
	marshal, err := json.Marshal(peers)
	if err != nil {
		logger.Error("Failed to marshal peers list")
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		logger.Error("Failed to write peers to response")
		return
	}
}

func writeErrorToResponse(w http.ResponseWriter, errs []error) {
	marshal, err := json.Marshal(errs)
	if err != nil {
		logger.Error("Failed to marshal error")
		return
	}
	w.WriteHeader(500)
	_, err = w.Write(marshal)
	if err != nil {
		logger.Error("Failed to write errors to response")
		return
	}
}

func peersToNgPeers(peers []tendermint.Peer) []NgPeer {
	var res []NgPeer

	for _, p := range peers {
		np := NgPeer{}
		np.NodeInfo = p.NodeInfo()
		if p.Get("lastValue") != nil {
			np.LastValue = p.Get("lastValue").(int)
		}
		if p.Get("status") != nil {
			np.Status = p.Get("status").(string)
		}
		if p.Get("error") != nil {
			np.Error = p.Get("error").(string)
		}
		res = append(res, np)
	}

	return res
}
