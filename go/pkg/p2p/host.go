// libp2p host initialization and management
package p2p

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

// Host manages the libp2p host and DHT
type Host struct {
	host   host.Host
	dht    *dht.IpfsDHT
	log    *logger.Logger
	config P2PConfig
}

// P2PConfig is imported from manager.go (config.P2PConfig)
// We'll use it directly from the Manager

// NewHost creates a new libp2p host with DHT
func NewHost(ctx context.Context, port int, bootstrapPeers []string, maxPeers int, log *logger.Logger) (*Host, error) {
	// Generate or load Ed25519 private key
	// TODO: In production, load from persistent storage
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Create listen addresses
	listenAddrs := []multiaddr.Multiaddr{}

	// TCP listener
	tcpAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP multiaddr: %w", err)
	}
	listenAddrs = append(listenAddrs, tcpAddr)

	// QUIC listener (UDP-based, more efficient)
	quicAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port))
	if err != nil {
		log.Warn("Failed to create QUIC multiaddr, skipping QUIC transport")
	} else {
		listenAddrs = append(listenAddrs, quicAddr)
	}

	// Connection manager (limits peer count)
	connMgr, err := connmgr.NewConnManager(
		maxPeers/2,  // Low water mark
		maxPeers,    // High water mark
		connmgr.WithGracePeriod(0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	// Create DHT (will be set during host creation)
	var dhtInstance *dht.IpfsDHT

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.ConnectionManager(connMgr),
		libp2p.NATPortMap(),                    // Enable NAT port mapping
		libp2p.EnableAutoRelay(),               // Enable auto relay for NAT traversal
		libp2p.EnableNATService(),              // Help other peers with NAT traversal
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			// Initialize DHT in server mode
			dhtInstance, err = dht.New(ctx, h, dht.Mode(dht.ModeAutoServer))
			return dhtInstance, err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	log.WithFields(logger.Fields{
		"peer_id": h.ID().String(),
		"addrs":   h.Addrs(),
	}).Info("libp2p host created")

	// Bootstrap DHT
	if err := dhtInstance.Bootstrap(ctx); err != nil {
		h.Close()
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	hostWrapper := &Host{
		host: h,
		dht:  dhtInstance,
		log:  log,
	}

	// Connect to bootstrap peers
	if len(bootstrapPeers) > 0 {
		log.WithField("count", len(bootstrapPeers)).Info("Connecting to bootstrap peers")
		hostWrapper.ConnectToBootstrapPeers(ctx, bootstrapPeers)
	} else {
		log.Warn("No bootstrap peers configured - node will not discover peers")
	}

	return hostWrapper, nil
}

// ConnectToBootstrapPeers connects to bootstrap nodes
func (h *Host) ConnectToBootstrapPeers(ctx context.Context, bootstrapPeers []string) {
	for _, peerAddr := range bootstrapPeers {
		maddr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			h.log.WithFields(logger.Fields{
				"addr":  peerAddr,
				"error": err,
			}).Warn("Invalid bootstrap peer address")
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			h.log.WithFields(logger.Fields{
				"addr":  peerAddr,
				"error": err,
			}).Warn("Failed to parse peer info")
			continue
		}

		if err := h.host.Connect(ctx, *peerInfo); err != nil {
			h.log.WithFields(logger.Fields{
				"peer_id": peerInfo.ID.String(),
				"error":   err,
			}).Warn("Failed to connect to bootstrap peer")
		} else {
			h.log.WithFields(logger.Fields{
				"peer_id": peerInfo.ID.String(),
				"addrs":   peerInfo.Addrs,
			}).Info("Connected to bootstrap peer")
		}
	}
}

// ID returns the host's peer ID
func (h *Host) ID() peer.ID {
	return h.host.ID()
}

// Addrs returns the host's listen addresses
func (h *Host) Addrs() []multiaddr.Multiaddr {
	return h.host.Addrs()
}

// GetHost returns the underlying libp2p host
func (h *Host) GetHost() host.Host {
	return h.host
}

// GetDHT returns the DHT instance
func (h *Host) GetDHT() *dht.IpfsDHT {
	return h.dht
}

// PeerCount returns number of connected peers
func (h *Host) PeerCount() int {
	return len(h.host.Network().Peers())
}

// ConnectedPeers returns list of connected peer IDs
func (h *Host) ConnectedPeers() []peer.ID {
	return h.host.Network().Peers()
}

// FindPeer finds a peer in the DHT
func (h *Host) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	return h.dht.FindPeer(ctx, peerID)
}

// Close closes the host and DHT
func (h *Host) Close() error {
	if err := h.dht.Close(); err != nil {
		h.log.WithError(err).Error("Failed to close DHT")
	}
	return h.host.Close()
}
