package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mindovermachine-dev/group-signup/internal/store"
)

func TestMemory(t *testing.T) {
	ctx := context.Background()
	m := store.NewMemory()

	if _, err := m.Get(ctx, "news", "a@example.com"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("Get on empty store: err = %v, want ErrNotFound", err)
	}

	want := store.Subscriber{Email: "a@example.com", List: "news", Status: store.Pending}
	if err := m.Put(ctx, want); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := m.Get(ctx, "news", "a@example.com")
	if err != nil || got != want {
		t.Fatalf("Get = %+v, %v; want %+v", got, err, want)
	}

	if _, err := m.Get(ctx, "other", "a@example.com"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("lists must be isolated: err = %v", err)
	}
}
