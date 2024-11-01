package shrinkmap

import (
	"sync"
	"time"
)

// Metrics tracks performance metrics of the map
type Metrics struct {
	mu                  sync.RWMutex
	totalShrinks        int64
	lastShrinkDuration  time.Duration
	totalItemsProcessed int64
	peakSize            int32
}

func (m *Metrics) TotalShrinks() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalShrinks
}

func (m *Metrics) LastShrinkDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastShrinkDuration
}

func (m *Metrics) TotalItemsProcessed() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalItemsProcessed
}

func (m *Metrics) PeakSize() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.peakSize
}
