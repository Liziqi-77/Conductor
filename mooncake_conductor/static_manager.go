// Copyright 2025 AIBrix Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build zmq
// +build zmq

package kvevent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/vllm-project/aibrix/pkg/cache/kvcache"
	"github.com/vllm-project/aibrix/pkg/utils"
)

// StaticManager manages KV event subscriptions for a fixed set of services.
type StaticManager struct {
	// Dependencies
	syncProvider SyncIndexProvider

	// Configuration
	services []ServiceConfig

	// Subscriber management
	// Using utils.SyncMap for type safety with Generics
	subscribers utils.SyncMap[string, *kvcache.ZMQClient]

	// Lifecycle management
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
	stopped bool
}

// NewStaticManager creates a new static KV event manager.
func NewStaticManager(
	services []ServiceConfig,
	syncProvider SyncIndexProvider,
) *StaticManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &StaticManager{
		services:     services,
		syncProvider: syncProvider,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start initializes the manager and establishes subscriptions for all configured services.
func (m *StaticManager) Start() error {
	klog.Info("Starting Static KV Event Manager...")

	// 1. Verify SyncIndexer availability
	initCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	if _, err := m.syncProvider.GetSyncIndexer(initCtx); err != nil {
		return fmt.Errorf("sync indexer not ready: %w", err)
	}

	// 2. Subscribe to all services concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(m.services))

	for _, svc := range m.services {
		wg.Add(1)
		go func(service ServiceConfig) {
			defer wg.Done()
			if err := m.subscribeToService(service); err != nil {
				klog.Errorf("Failed to initiate subscription for %s service %s (%s): %v",
					service.Type, service.Name, service.IP, err)
				errChan <- fmt.Errorf("failed to subscribe to %s: %w", service.Name, err)
			}
		}(svc)
	}

	wg.Wait()
	close(errChan)

	failureCount := len(errChan)
	successCount := len(m.services) - failureCount
	klog.Infof("Static KV Event Manager started. Subscriptions: %d success, %d failed (will retry internally)",
		successCount, failureCount)

	return nil
}

// Stop gracefully shuts down the manager and all subscriptions.
func (m *StaticManager) Stop() {
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return
	}
	m.stopped = true
	m.mu.Unlock()

	klog.Info("Stopping Static KV Event Manager")

	// 1. Cancel context
	m.cancel()

	// 2. Stop all ZMQ clients
	m.subscribers.Range(func(key string, client *kvcache.ZMQClient) bool {
		client.Stop()
		klog.Infof("Stopped subscription for %s", key)
		return true
	})
}

// subscribeToService establishes a ZMQ subscription for a single service.
func (m *StaticManager) subscribeToService(svc ServiceConfig) error {
	if _, exists := m.subscribers.Load(svc.Name); exists {
		return nil
	}

	// Create handler instance directly.
	// The implementation of staticEventHandler is in static_handler.go
	handler := &staticEventHandler{
		manager:   m,
		svcName:   svc.Name,
		modelName: svc.ModelName,
		loraID:    svc.LoraID,
	}

	// Configure ZMQ Client
	zmqConfig := &kvcache.ZMQClientConfig{
		PodKey:         svc.Name,
		PodIP:          svc.IP,
		PubPort:        svc.Port,
		ModelName:      svc.ModelName,
		PollTimeout:    100 * time.Millisecond,
		ReplayTimeout:  5 * time.Second,
		ReconnectDelay: 1 * time.Second,
		RouterPort:     svc.Port + 1,
	}

	// Create and start client
	client := kvcache.NewZMQClient(zmqConfig, handler)
	if err := client.Start(); err != nil {
		return fmt.Errorf("failed to start ZMQ client: %w", err)
	}

	m.subscribers.Store(svc.Name, client)
	klog.Infof("Successfully subscribed to %s service: %s at %s:%d",
		svc.Type, svc.Name, svc.IP, svc.Port)

	return nil
}
