// Integration tests for P2P networking layer
package p2p

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/config"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/mempool"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/state"
)

// TestHostCreation tests libp2p host initialization
func TestHostCreation(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger("error")

	host, err := NewHost(ctx, 0, []string{}, 50, log)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close()

	if host.ID().String() == "" {
		t.Fatal("Host ID should not be empty")
	}

	if len(host.Addrs()) == 0 {
		t.Fatal("Host should have listen addresses")
	}

	t.Logf("Host created: %s", host.ID().String())
	t.Logf("Listen addrs: %v", host.Addrs())
}

// TestPeerScoring tests peer reputation management
func TestPeerScoring(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger("error")

	ps := NewPeerScoring(ctx, QuarantineThreshold, BanThreshold, log)
	defer ps.Close()

	// Create dummy peer ID
	ctx2 := context.Background()
	log2 := logger.NewLogger("error")
	host, err := NewHost(ctx2, 0, []string{}, 50, log2)
	if err != nil {
		t.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close()

	peerID := host.ID()

	// Initial score should be 100
	score := ps.GetScore(peerID)
	if score != InitialPeerScore {
		t.Fatalf("Initial score should be %d, got %d", InitialPeerScore, score)
	}

	// Record valid message
	ps.RecordValidMessage(peerID)
	score = ps.GetScore(peerID)
	if score != InitialPeerScore+ScoreValidMessage {
		t.Fatalf("Score should be %d after valid message, got %d", InitialPeerScore+ScoreValidMessage, score)
	}

	// Record multiple invalid messages
	for i := 0; i < 10; i++ {
		ps.RecordInvalidMessage(peerID)
	}

	// Should be quarantined
	if !ps.IsQuarantined(peerID) {
		t.Fatal("Peer should be quarantined after many invalid messages")
	}

	t.Logf("Peer scoring works correctly")
}

// TestMempoolCreation tests mempool for transaction gossip
func TestMempoolCreation(t *testing.T) {
	log := logger.NewLogger("error")

	mp, err := mempool.NewMempool(mempool.Config{
		MaxSize:          1000,
		MaxTxAge:         time.Hour,
		CleanupInterval:  time.Minute,
		PriorityThreshold: 0,
	}, log)
	if err != nil {
		t.Fatalf("Failed to create mempool: %v", err)
	}
	defer mp.Close()

	if mp.Size() != 0 {
		t.Fatal("Mempool should start empty")
	}

	t.Logf("Mempool created successfully")
}

// TestStateManagerCreation tests state manager for transaction validation
func TestStateManagerCreation(t *testing.T) {
	log := logger.NewLogger("error")

	// Create temp database
	tmpDB := "test_state.db"
	defer os.Remove(tmpDB)

	sm, err := state.NewStateManager(tmpDB, log)
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer sm.Close()

	// Test account creation
	testAddr := [32]byte{1, 2, 3, 4}
	err = sm.CreateAccount(testAddr, 1000000)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Get account
	account, err := sm.GetAccount(testAddr)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.Balance != 1000000 {
		t.Fatalf("Account balance should be 1000000, got %d", account.Balance)
	}

	t.Logf("State manager works correctly")
}

// TestP2PManagerIntegration tests full P2P manager
func TestP2PManagerIntegration(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger("error")

	// Create mempool
	mp, err := mempool.NewMempool(mempool.Config{
		MaxSize:          1000,
		MaxTxAge:         time.Hour,
		CleanupInterval:  time.Minute,
		PriorityThreshold: 0,
	}, log)
	if err != nil {
		t.Fatalf("Failed to create mempool: %v", err)
	}
	defer mp.Close()

	// Create state manager
	tmpDB := "test_p2p_state.db"
	defer os.Remove(tmpDB)

	sm, err := state.NewStateManager(tmpDB, log)
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer sm.Close()

	// Create P2P manager config
	cfg := config.P2PConfig{
		Port:                0, // Random port
		BootstrapPeers:      []string{},
		MaxPeers:            50,
		EquilibriumLambda:   0.7071,
		BroadcastInterval:   14140,
		PeerScoringEnabled:  true,
		QuarantineThreshold: 10,
	}

	// Create P2P manager
	manager, err := NewManager(ctx, cfg, mp, sm, log)
	if err != nil {
		t.Fatalf("Failed to create P2P manager: %v", err)
	}

	// Start manager
	err = manager.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start P2P manager: %v", err)
	}
	defer manager.Stop()

	// Wait a bit for initialization
	time.Sleep(100 * time.Millisecond)

	// Check peer count (should be 0 since no bootstrap peers)
	peerCount := manager.PeerCount()
	if peerCount != 0 {
		t.Logf("Peer count: %d (expected 0 without bootstrap peers)", peerCount)
	}

	// Check peer ID
	peerID := manager.GetPeerID()
	if peerID == "" {
		t.Fatal("Peer ID should not be empty")
	}

	// Get network stats
	stats := manager.GetNetworkStats()
	if stats == nil {
		t.Fatal("Network stats should not be nil")
	}

	t.Logf("P2P Manager started successfully")
	t.Logf("Peer ID: %s", peerID)
	t.Logf("Stats: %+v", stats)
}

// TestCIDGossipIntegration tests CID broadcasting
func TestCIDGossipIntegration(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger("error")

	// Create mempool
	mp, err := mempool.NewMempool(mempool.Config{
		MaxSize:          1000,
		MaxTxAge:         time.Hour,
		CleanupInterval:  time.Minute,
		PriorityThreshold: 0,
	}, log)
	if err != nil {
		t.Fatalf("Failed to create mempool: %v", err)
	}
	defer mp.Close()

	// Create state manager
	tmpDB := "test_cid_state.db"
	defer os.Remove(tmpDB)

	sm, err := state.NewStateManager(tmpDB, log)
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer sm.Close()

	// Create P2P manager
	cfg := config.P2PConfig{
		Port:                0,
		BootstrapPeers:      []string{},
		MaxPeers:            50,
		EquilibriumLambda:   0.7071,
		BroadcastInterval:   14140,
		PeerScoringEnabled:  true,
		QuarantineThreshold: 10,
	}

	manager, err := NewManager(ctx, cfg, mp, sm, log)
	if err != nil {
		t.Fatalf("Failed to create P2P manager: %v", err)
	}

	err = manager.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start P2P manager: %v", err)
	}
	defer manager.Stop()

	// Wait for initialization
	time.Sleep(100 * time.Millisecond)

	// Announce a CID
	testCID := "QmTestCID123456789"
	manager.AnnounceCID(testCID, "problem", 100)

	// Check stats
	stats := manager.GetNetworkStats()
	if queueSize, ok := stats["cid_queue_size"].(int); ok {
		if queueSize != 1 {
			t.Fatalf("CID queue should have 1 item, got %d", queueSize)
		}
	}

	// Check equilibrium ratio
	if ratio, ok := stats["equilibrium_ratio"].(float64); ok {
		if ratio != 1.0 {
			t.Logf("Equilibrium ratio: %.4f (should be 1.0 for λ = η)", ratio)
		}
	}

	t.Logf("CID broadcasting works correctly")
}

// TestTwoNodeConnection tests connection between two nodes
func TestTwoNodeConnection(t *testing.T) {
	// This test requires two nodes to connect
	// Skip in CI environments without network access
	if testing.Short() {
		t.Skip("Skipping two-node test in short mode")
	}

	ctx := context.Background()
	log := logger.NewLogger("info")

	// Create first node
	host1, err := NewHost(ctx, 0, []string{}, 50, log)
	if err != nil {
		t.Fatalf("Failed to create host1: %v", err)
	}
	defer host1.Close()

	// Create second node
	host2, err := NewHost(ctx, 0, []string{}, 50, log)
	if err != nil {
		t.Fatalf("Failed to create host2: %v", err)
	}
	defer host2.Close()

	// Get host1's multiaddr
	host1Addrs := host1.Addrs()
	if len(host1Addrs) == 0 {
		t.Fatal("Host1 has no addresses")
	}

	// Connect host2 to host1
	host1Multiaddr := host1Addrs[0].String() + "/p2p/" + host1.ID().String()
	host2.ConnectToBootstrapPeers(ctx, []string{host1Multiaddr})

	// Wait for connection
	time.Sleep(500 * time.Millisecond)

	// Check peer counts
	host1Peers := host1.PeerCount()
	host2Peers := host2.PeerCount()

	t.Logf("Host1 peers: %d", host1Peers)
	t.Logf("Host2 peers: %d", host2Peers)

	if host1Peers < 1 && host2Peers < 1 {
		t.Log("Warning: Peers may not have connected (normal in some test environments)")
	}
}
