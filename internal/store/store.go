// Package store persists newsletter subscribers.
package store

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrNotFound is returned when a subscriber does not exist.
var ErrNotFound = errors.New("store: not found")

// Status is the lifecycle state of a subscription.
type Status string

const (
	Pending      Status = "pending"
	Confirmed    Status = "confirmed"
	Unsubscribed Status = "unsubscribed"
)

// Subscriber is one email address on one list.
type Subscriber struct {
	Email       string    `firestore:"email"`
	List        string    `firestore:"list"`
	Status      Status    `firestore:"status"`
	CreatedAt   time.Time `firestore:"createdAt"`
	ConfirmedAt time.Time `firestore:"confirmedAt"`
}

// Store is the persistence boundary for subscribers.
type Store interface {
	Get(ctx context.Context, list, email string) (Subscriber, error)
	Put(ctx context.Context, s Subscriber) error
}

// Memory is an in-memory Store for tests and local development.
type Memory struct {
	mu   sync.Mutex
	subs map[[2]string]Subscriber
}

// NewMemory returns an empty Memory store.
func NewMemory() *Memory {
	return &Memory{subs: make(map[[2]string]Subscriber)}
}

func (m *Memory) Get(_ context.Context, list, email string) (Subscriber, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.subs[[2]string{list, email}]
	if !ok {
		return Subscriber{}, ErrNotFound
	}
	return s, nil
}

func (m *Memory) Put(_ context.Context, s Subscriber) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs[[2]string{s.List, s.Email}] = s
	return nil
}
