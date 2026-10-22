package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const subscribersCollection = "subscribers"

// Firestore is a Store backed by Cloud Firestore.
type Firestore struct {
	client *firestore.Client
}

// NewFirestore connects to Firestore in the given project.
func NewFirestore(ctx context.Context, project string) (*Firestore, error) {
	c, err := firestore.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	return &Firestore{client: c}, nil
}

// Close releases the underlying connection.
func (f *Firestore) Close() error { return f.client.Close() }

func (f *Firestore) Get(ctx context.Context, list, email string) (Subscriber, error) {
	snap, err := f.doc(list, email).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return Subscriber{}, ErrNotFound
	}
	if err != nil {
		return Subscriber{}, fmt.Errorf("firestore get: %w", err)
	}
	var s Subscriber
	if err := snap.DataTo(&s); err != nil {
		return Subscriber{}, fmt.Errorf("firestore decode: %w", err)
	}
	return s, nil
}

func (f *Firestore) Put(ctx context.Context, s Subscriber) error {
	if _, err := f.doc(s.List, s.Email).Set(ctx, s); err != nil {
		return fmt.Errorf("firestore set: %w", err)
	}
	return nil
}

// The hashed ID makes (list, email) unique and keeps addresses out of paths.
func (f *Firestore) doc(list, email string) *firestore.DocumentRef {
	sum := sha256.Sum256([]byte(list + "\x00" + email))
	return f.client.Collection(subscribersCollection).Doc(hex.EncodeToString(sum[:]))
}
