package conversation

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrConversationNotFound = errors.New("conversation not found")

type Store interface {
	Get(context.Context, string) (*State, error)
	Save(context.Context, string, *State) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]*State
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]*State)}
}

func (s *MemoryStore) Get(_ context.Context, key string) (*State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.items[key]
	if !ok {
		return nil, ErrConversationNotFound
	}
	cp := *state
	cp.LastProjectIDs = append([]int(nil), state.LastProjectIDs...)
	cp.LastContractIDs = append([]int(nil), state.LastContractIDs...)
	cp.LastTaskUUIDs = append([]string(nil), state.LastTaskUUIDs...)
	cp.LastArticleLabels = append([]string(nil), state.LastArticleLabels...)
	return &cp, nil
}

func (s *MemoryStore) Save(_ context.Context, key string, state *State) error {
	if state == nil {
		return fmt.Errorf("conversation state is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *state
	cp.LastProjectIDs = append([]int(nil), state.LastProjectIDs...)
	cp.LastContractIDs = append([]int(nil), state.LastContractIDs...)
	cp.LastTaskUUIDs = append([]string(nil), state.LastTaskUUIDs...)
	cp.LastArticleLabels = append([]string(nil), state.LastArticleLabels...)
	s.items[key] = &cp
	return nil
}
