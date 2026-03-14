package ingest

import (
	"sync"

	"github.com/eng-graph/eng-graph/internal/source"
	"github.com/eng-graph/eng-graph/internal/storage"
)

type Deduplicator struct {
	store *storage.SQLiteStore
	mu    sync.Mutex
	seen  map[string]bool
}

func NewDeduplicator(store *storage.SQLiteStore) *Deduplicator {
	return &Deduplicator{
		store: store,
		seen:  make(map[string]bool),
	}
}

func (d *Deduplicator) IsDuplicate(profileName string, dp source.DataPoint) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.seen[dp.ID] {
		return true
	}

	exists, err := d.store.Exists(profileName, dp.ID)
	if err == nil && exists {
		d.seen[dp.ID] = true
		return true
	}

	d.seen[dp.ID] = true
	return false
}
