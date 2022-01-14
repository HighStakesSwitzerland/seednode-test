package tendermint

import (
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
	"os"

	"fmt"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
	tmp2p "github.com/tendermint/tendermint/proto/tendermint/p2p"
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
	GetChannels() []*ChannelDescriptor
	InitPeer(peer Peer) Peer
	AddPeer(peer Peer)
	RemovePeer(peer Peer, reason interface{})
	Receive(chID byte, peer Peer, msgBytes []byte)
}

type SeedNodeReactor struct {
	p2p.BaseReactor
	Switch *Switch
}

// NewReactor creates new PEX reactor.
func NewReactor() *SeedNodeReactor {
	r := &SeedNodeReactor{}
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
	// logger.Debug(fmt.Sprintf("Requesting addrs from %s", p.ID()))
	p.Send(pex.PexChannel, mustEncode(&tmp2p.PexRequest{}))
}

func (br *SeedNodeReactor) Receive(chID byte, src Peer, msgBytes []byte) {
	msg, err := decodeMsg(msgBytes)

	if err != nil {
		logger.Error("Error decoding message", "src", src, "chId", chID, "err", err)
		br.Switch.StopPeerForError(src, err)
		src.Set("warning", "down")
		return
	}

	switch msg := msg.(type) {
	case *tmp2p.PexRequest:
		// we don't care, and it should never happen
		logger.Error("Received a PexRequest")
		break

	case *tmp2p.PexAddrs:
		// We asked for addresses, add them to the source peer
		addrs, err := p2p.NetAddressesFromProto(msg.Addrs)
		logger.Info(fmt.Sprintf("We got %d addresses from %s", len(addrs), src.NodeInfo().ID()))

		peer := br.Switch.persistentPeers.Get(src.ID())
		if peer == nil {
			panic("persistentPeers does not contain the source peer!?")
		}
		src.Set("lastValue", len(addrs))
		peer.GetMetrics().SeedReceivePeers.Set(float64(len(addrs)))

		if err != nil {
			logger.Error(err.Error())
			return
		}

	default:
		logger.Error(fmt.Sprintf("Unknown message type %T", msg))
	}
}

// DialPeers Dials a peer and retrieve the resulting seed list
func DialPeers(peers []string, r *SeedNodeReactor) []error {
	addrs, errors := p2p.NewNetAddressStrings(peers)
	if len(errors) > 0 {
		logger.Error("Detected invalid peer(s)", "error", errors)
		// TODO reactor.Switch.GetPersistentPeers().Has()
		// find missing peers with intersection
	}

	for _, netAddr := range addrs {

		err := r.Switch.DialPeerWithAddress(netAddr)
		if err != nil {
			logger.Error(err.Error(), "addr", netAddr)
			peer := r.Switch.persistentPeers.Get(netAddr.ID)
			if peer != nil {
				peer.Set("status", "warning")
				peer.Set("error", err.Error())
			}
		}
	}

	return errors
}

// GetChannels implements Reactor
func (br *SeedNodeReactor) GetChannels() []*ChannelDescriptor {
	return []*ChannelDescriptor{
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
