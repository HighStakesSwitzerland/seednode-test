package tendermint

import (
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/p2p/conn"
	"os"

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

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "seednode")
)

type Reactor interface {
	service.Service // Start, Stop
	SetSwitch(*Switch)
	GetChannels() []*conn.ChannelDescriptor
	InitPeer(peer Peer) Peer
	AddPeer(peer Peer)
	RemovePeer(peer Peer, reason interface{})
	Receive(chID byte, peer Peer, msgBytes []byte)
}

type SeedNodeReactor struct {
	p2p.BaseReactor
	Switch *Switch

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
	r.BaseReactor = *NewBaseReactor("PEX", r)
	return r
}

func NewBaseReactor(name string, impl Reactor) *p2p.BaseReactor {
	return &p2p.BaseReactor{
		BaseService: *service.NewBaseService(nil, name, impl),
		Switch:      nil,
	}
}

// RequestAddrs asks peer for more addresses if we do not already have a
// request out for this peer.
func (br *SeedNodeReactor) RequestAddrs(p Peer) {
	id := string(p.ID())
	if br.requestsSent.Has(id) {
		return
	}
	br.Logger.Debug("Request addrs", "from", p)
	br.requestsSent.Set(id, struct{}{})
	p.Send(pex.PexChannel, mustEncode(&tmp2p.PexRequest{}))
}

func (br *SeedNodeReactor) Receive(chID byte, src Peer, msgBytes []byte) {
	msg, err := decodeMsg(msgBytes)

	if err != nil {
		br.Logger.Error("Error decoding message", "src", src, "chId", chID, "err", err)
		br.Switch.StopPeerForError(src, err)
		return
	}
	br.Logger.Debug("Received message ", "src", src, "chId", chID, "msg", msg)

	switch msg := msg.(type) {
	case *tmp2p.PexRequest:
		// we don't care, and it should never happen
		logger.Error("Received a PexRequest")
		break

	case *tmp2p.PexAddrs:
		// We asked for addresses, return them to the front and forget them
		_, err := p2p.NetAddressesFromProto(msg.Addrs)
		if err != nil {
			logger.Error(err.Error())
			return
		}

	default:
		br.Logger.Error(fmt.Sprintf("Unknown message type %T", msg))
	}
}

// DialPeer Dials a peer and retrieve the resulting seed list
func DialPeer(addrs []*p2p.NetAddress, srcAddr *p2p.NetAddress, r *SeedNodeReactor) error {
	for _, netAddr := range addrs {
		logger.Info("Will dial address", "addr", netAddr, "seed", srcAddr)
		err := r.Switch.DialPeerWithAddress(netAddr)
		if err != nil {
			r.Logger.Error(err.Error(), "addr", netAddr)
		}
	}

	return nil
}

// GetChannels implements Reactor
func (br *SeedNodeReactor) GetChannels() []*conn.ChannelDescriptor {
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
func (br *SeedNodeReactor) AddPeer(p Peer) {
	br.RequestAddrs(p)
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

func (br *SeedNodeReactor) SetSwitch(sw *Switch) {
	br.Switch = sw
}

func (*SeedNodeReactor) RemovePeer(Peer, interface{}) {}
func (*SeedNodeReactor) InitPeer(peer Peer) Peer      { return peer }
