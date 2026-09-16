package cacheadapter

import (
	"testing"
)

func TestNoopCacheAdapter(t *testing.T) {
	adapter := &NoopCacheAdapter{}

	// Test Search
	ans, hit := adapter.Search([]float32{0.1, 0.2}, 0.8)
	if ans != "" || hit != false {
		t.Errorf("Expected Search to return ('', false), got ('%s', %v)", ans, hit)
	}

	// Test Store
	err := adapter.Store("q", "a")
	if err != nil {
		t.Errorf("Expected Store to return nil, got %v", err)
	}

	// Test Count
	count := adapter.Count()
	if count != 0 {
		t.Errorf("Expected Count to return 0, got %d", count)
	}
}
