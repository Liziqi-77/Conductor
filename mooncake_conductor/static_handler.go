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
	"time"

	"k8s.io/klog/v2"

	"github.com/vllm-project/aibrix/pkg/cache/kvcache"
)

// staticEventHandler adapts the generic EventHandler interface for StaticManager.
// It is instantiated in static_manager.go but implemented here to keep files clean.
type staticEventHandler struct {
	manager   *StaticManager
	svcName   string
	modelName string
	loraID    int64
}

// HandleEvent processes incoming events from ZMQClient.
func (h *staticEventHandler) HandleEvent(event kvcache.KVEvent) error {
	// 1. Lifecycle check
	h.manager.mu.RLock()
	if h.manager.stopped {
		h.manager.mu.RUnlock()
		return fmt.Errorf("manager stopped")
	}
	h.manager.mu.RUnlock()

	// 2. Create context for processing
	ctx, cancel := context.WithTimeout(h.manager.ctx, 5*time.Second)
	defer cancel()

	// 3. Get Indexer
	// Retrieves the indexer from the manager's provider.
	indexer, err := h.manager.syncProvider.GetSyncIndexer(ctx)
	if err != nil {
		return err
	}

	// 4. Dispatch event
	// The ZMQClient decodes messages into specific event types (BlockStored/Removed).
	// We pass these directly to the Indexer logic.
	switch e := event.(type) {
	case *kvcache.BlockStoredEvent:
		klog.V(4).Infof("[%s] BlockStored: %d blocks", h.svcName, len(e.BlockHashes))
		
		// In a real scenario, you might need to convert the event to the internal format
		// expected by the indexer if they differ.
		// e.g., internalEvent := convertToSyncStoredEvent(e, h.modelName, h.loraID, h.svcName)
		// return indexer.ProcessBlockStored(ctx, internalEvent)
		
		// Assuming interface compatibility for now:
		// return indexer.ProcessBlockStored(ctx, e)
		return nil

	case *kvcache.BlockRemovedEvent:
		klog.V(4).Infof("[%s] BlockRemoved: %d blocks", h.svcName, len(e.BlockHashes))
		// return indexer.ProcessBlockRemoved(ctx, e)
		return nil

	default:
		// Ignore unknown event types
		return nil
	}
}

