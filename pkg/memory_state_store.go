package pkg

import (
	"fmt"
	"sync"
)

type MemoryStateStore struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		store: make(map[string]string),
	}
}

func (m *MemoryStateStore) Get(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName) (string, error) {
	m.mu.RLock()
	key := fmt.Sprintf("%s-%s-%s", locationKey, instanceKey, operationName)
	content, exists := m.store[key]
	m.mu.RUnlock()

	if exists {
		return content, nil
	}

	return "", nil
}

func (m *MemoryStateStore) Set(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s-%s-%s", locationKey, instanceKey, operationName)
	m.store[key] = value
	return nil
}

func (m *MemoryStateStore) Close() error {
	return nil
}
