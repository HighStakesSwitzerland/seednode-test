package seednode

import (
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/p2p/conn"

	"fmt"
	"github.com/tendermint/tendermint/libs/cmap"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
	tmp2p "github.com/tendermint/tendermint/proto/tendermint/p2p"
	"sync"
	"time"
)

/**
 * Copied from pex.pex_reactor.go and adapted to our needs
 */

type SeedNodeReactor struct {
	p2p.BaseReactor

	book              pex.AddrBook
	config            *pex.ReactorConfig
	ensurePeersPeriod time.Duration // TODO: should go in the config

	// maps to prevent abuse
	requestsSent         *cmap.CMap // ID->struct{}: unanswered send requests
	lastReceivedRequests *cmap.CMap // ID->time.Time: last time peer requested from us

	seedAddrs []*p2p.NetAddress

	attemptsToDial sync.Map // address (string) -> {number of attempts (int), last time dialed (time.Time)}
}

// NewReactor creates new PEX reactor.
func NewReactor(b pex.AddrBook, config *pex.ReactorConfig) *SeedNodeReactor {
	r := &SeedNodeReactor{
		book:                 b,
		config:               config,
		ensurePeersPeriod:    30 * time.Second,
		requestsSent:         cmap.NewCMap(),
		lastReceivedRequests: cmap.NewCMap(),
	}
	r.BaseReactor = *p2p.NewBaseReactor("PEX", r)
	return r
}

// RequestAddrs asks peer for more addresses if we do not already have a
// request out for this peer.
func (r *SeedNodeReactor) RequestAddrs(p pex.Peer) {
	id := string(p.ID())
	if r.requestsSent.Has(id) {
		return
	}
	r.Logger.Debug("Request addrs", "from", p)
	r.requestsSent.Set(id, struct{}{})
	p.Send(pex.PexChannel, mustEncode(&tmp2p.PexRequest{}))
}

func (r *SeedNodeReactor) Receive(chID byte, src pex.Peer, msgBytes []byte) {
	msg, err := decodeMsg(msgBytes)

	if err != nil {
		r.Logger.Error("Error decoding message", "src", src, "chId", chID, "err", err)
		r.Switch.StopPeerForError(src, err)
		return
	}
	r.Logger.Debug("Received message ", "src", src, "chId", chID, "msg", msg)

	switch msg := msg.(type) {
	case *tmp2p.PexRequest:
		// we don't care, and it should never happen
		break

	case *tmp2p.PexAddrs:
		// We asked for addresses, return them to the front and forget them
		_, err := p2p.NetAddressesFromProto(msg.Addrs)
		if err != nil {
			logger.Error(err.Error())
			return
		}

	default:
		r.Logger.Error(fmt.Sprintf("Unknown message type %T", msg))
	}
}

// DialPeer Dials a peer and retrieve the resulting seed list
func DialPeer(addrs []*p2p.NetAddress, src p2p.Peer, r *SeedNodeReactor) error {
	srcAddr, err := src.NodeInfo().NetAddress()
	if err != nil {
		return err
	}

	for _, netAddr := range addrs {
		logger.Info("Will dial address, which came from seed", "addr", netAddr, "seed", srcAddr)
		err := r.Switch.DialPeerWithAddress(netAddr)
		if err != nil {
			r.Logger.Error(err.Error(), "addr", netAddr)
		}
	}

	return nil
}

// GetChannels implements Reactor
func (r *SeedNodeReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{
		{
			ID:                  pex.PexChannel,
			Priority:            1,
			SendQueueCapacity:   10,
			RecvMessageCapacity: 256 * 250,
		},
	}
}

// AddPeer normally implements Reactor by adding peer to the address book (if inbound)
// or by requesting more addresses (if outbound).
// This version only request addressed
func (r *SeedNodeReactor) AddPeer(p pex.Peer) {
	r.RequestAddrs(p)
}

// mustEncode proto encodes a tmp2p.Message
func mustEncode(pb proto.Message) []byte {
	msg := tmp2p.Message{}
	switch pb := pb.(type) {
	case *tmp2p.PexRequest:
		msg.Sum = &tmp2p.Message_PexRequest{PexRequest: pb}
	case *tmp2p.PexAddrs:
		msg.Sum = &tmp2p.Message_PexAddrs{PexAddrs: pb}
	default:
		panic(fmt.Sprintf("Unknown message type %T", pb))
	}

	bz, err := msg.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal %T: %w", pb, err))
	}
	return bz
}

func decodeMsg(bz []byte) (proto.Message, error) {
	pb := &tmp2p.Message{}

	err := pb.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	switch msg := pb.Sum.(type) {
	case *tmp2p.Message_PexRequest:
		return msg.PexRequest, nil
	case *tmp2p.Message_PexAddrs:
		return msg.PexAddrs, nil
	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
