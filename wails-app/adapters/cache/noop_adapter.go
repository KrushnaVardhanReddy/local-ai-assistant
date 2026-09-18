package cacheadapter

import (
	"wails-app/backend"
	"wails-app/core/ports/driven"
)

// NoopCacheAdapter is a no-op implementation for tests and offline mode.
// It never finds a cache hit and silently discards stores.
type NoopCacheAdapter struct{}

func (n *NoopCacheAdapter) Search(_ []float32, _ float64) (string, bool) { return "", false }
func (n *NoopCacheAdapter) Store(_, _ string) error                       { return nil }
func (n *NoopCacheAdapter) Count() int                                    { return 0 }
func (n *NoopCacheAdapter) GetAllItems() ([]backend.CacheItem, error)     { return nil, nil }

var _ driven.CachePort = (*NoopCacheAdapter)(nil)
