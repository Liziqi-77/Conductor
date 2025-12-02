//go:build zmq
// +build zmq

package kvevent

import (
	"context"
	"testing"
	"time"

	"github.com/vllm-project/aibrix/pkg/cache/kvcache"
	"k8s.io/klog/v2"
)

// -----------------------------------------------------------------------------
// Mocks for Testing
// -----------------------------------------------------------------------------

// MockSyncProvider implements SyncIndexProvider
type MockSyncProvider struct{}

func (m *MockSyncProvider) GetSyncIndexer(ctx context.Context) (SyncIndexer, error) {
	return &MockSyncIndexer{}, nil
}

// MockSyncIndexer implements SyncIndexer
type MockSyncIndexer struct{}

func (m *MockSyncIndexer) ProcessBlockStored(ctx context.Context, event interface{}) error {
	klog.Info("MockSyncIndexer: ProcessBlockStored called")
	return nil
}

func (m *MockSyncIndexer) ProcessBlockRemoved(ctx context.Context, event interface{}) error {
	klog.Info("MockSyncIndexer: ProcessBlockRemoved called")
	return nil
}

// -----------------------------------------------------------------------------
// Test Cases
// -----------------------------------------------------------------------------

func TestStaticManager_Lifecycle(t *testing.T) {
	// 1. Prepare Configuration
	services := []ServiceConfig{
		{
			Name:      "test-vllm-01",
			IP:        "127.0.0.1", // Localhost for testing
			Port:      55557,       // Random high port
			Type:      ServiceTypeVLLM,
			ModelName: "test-model",
		},
		{
			Name:      "test-mooncake-01",
			IP:        "127.0.0.1",
			Port:      55558,
			Type:      ServiceTypeMooncake,
			ModelName: "test-model-2",
		},
	}

	// 2. Create Dependencies
	provider := &MockSyncProvider{}

	// 3. Instantiate Manager
	// Now using the standard constructor without factory injection
	manager := NewStaticManager(services, provider)

	// 4. Test Start()
	// Start() will internally create staticEventHandler and ZMQClient for each service.
	t.Log("Starting Manager...")
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Verify internal state
	count := 0
	manager.subscribers.Range(func(key string, value *kvcache.ZMQClient) bool {
		t.Logf("Subscriber active for: %s", key)
		count++
		if value == nil {
			t.Errorf("Subscriber client is nil for %s", key)
		}
		return true
	})

	if count != len(services) {
		t.Errorf("Expected %d subscribers, got %d", len(services), count)
	}

	// 5. Simulate running for a bit
	time.Sleep(100 * time.Millisecond)

	// 6. Test Stop()
	t.Log("Stopping Manager...")
	manager.Stop()

	// Verify double stop doesn't panic
	manager.Stop()
	t.Log("Manager lifecycle test passed")
}

// TestStaticManager_SubscribeIdempotency verifies that calling subscribe multiple times works
func TestStaticManager_SubscribeIdempotency(t *testing.T) {
	services := []ServiceConfig{
		{Name: "dup-service", IP: "127.0.0.1", Port: 6000, Type: ServiceTypeVLLM},
	}

	mgr := NewStaticManager(services, &MockSyncProvider{})

	// First start
	if err := mgr.Start(); err != nil {
		t.Fatalf("First start failed: %v", err)
	}

	// Manually try to subscribe again (internal method test)
	err := mgr.subscribeToService(services[0])
	if err != nil {
		t.Errorf("Duplicate subscription request should not return error: %v", err)
	}

	mgr.Stop()
}
