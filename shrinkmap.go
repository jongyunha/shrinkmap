package shrinkmap

import (
	"sync"
	"sync/atomic"
	"time"
)

// ShrinkableMap provides a generic map structure with automatic shrinking capabilities
type ShrinkableMap[K comparable, V any] struct {
	mu             sync.RWMutex
	data           map[K]V
	itemCount      int32 // Use atomic operations
	deletedCount   int32 // Use atomic operations
	config         Config
	lastShrinkTime atomic.Value // Use atomic for time value
	metrics        *Metrics     // Use pointer to ensure atomic access
	shrinking      atomic.Bool  // Flag to track shrink operation
}

// New creates a new ShrinkableMap with the given configuration
func New[K comparable, V any](config Config) *ShrinkableMap[K, V] {
	sm := &ShrinkableMap[K, V]{
		data:    make(map[K]V, config.InitialCapacity),
		config:  config,
		metrics: &Metrics{},
	}

	sm.lastShrinkTime.Store(time.Now())

	if config.AutoShrinkEnabled {
		go sm.shrinkLoop()
	}
	return sm
}

// Set stores a key-value pair in the map
func (sm *ShrinkableMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	_, exists := sm.data[key]
	sm.data[key] = value
	sm.mu.Unlock()

	if !exists {
		atomic.AddInt32(&sm.itemCount, 1)
		sm.updateMetrics(1)
	}

	// Check if shrink is needed
	if sm.config.MaxMapSize > 0 && atomic.LoadInt32(&sm.itemCount) >= int32(sm.config.MaxMapSize) {
		sm.TryShrink()
	}
}

// Get retrieves the value associated with the given key
func (sm *ShrinkableMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	value, exists := sm.data[key]
	sm.mu.RUnlock()
	return value, exists
}

// Delete removes the entry for the given key
func (sm *ShrinkableMap[K, V]) Delete(key K) bool {
	sm.mu.Lock()
	_, exists := sm.data[key]
	if exists {
		delete(sm.data, key)
		atomic.AddInt32(&sm.deletedCount, 1)
	}
	sm.mu.Unlock()

	if exists && sm.config.AutoShrinkEnabled {
		sm.TryShrink()
	}
	return exists
}

// Len returns the current number of items in the map
func (sm *ShrinkableMap[K, V]) Len() int {
	return int(atomic.LoadInt32(&sm.itemCount) - atomic.LoadInt32(&sm.deletedCount))
}

// updateMetrics safely updates the metrics
func (sm *ShrinkableMap[K, V]) updateMetrics(processedItems int64) {
	sm.metrics.mu.Lock()
	defer sm.metrics.mu.Unlock()

	sm.metrics.totalItemsProcessed += processedItems
	currentSize := atomic.LoadInt32(&sm.itemCount)
	if currentSize > sm.metrics.peakSize {
		sm.metrics.peakSize = currentSize
	}
}

// GetMetrics returns a copy of the current metrics
func (sm *ShrinkableMap[K, V]) GetMetrics() Metrics {
	sm.metrics.mu.RLock()
	defer sm.metrics.mu.RUnlock()
	return Metrics{
		totalShrinks:        sm.metrics.totalShrinks,
		lastShrinkDuration:  sm.metrics.lastShrinkDuration,
		totalItemsProcessed: sm.metrics.totalItemsProcessed,
		peakSize:            sm.metrics.peakSize,
	}
}

// shouldShrink determines if the map should be shrunk based on current conditions
func (sm *ShrinkableMap[K, V]) shouldShrink() bool {
	itemCount := atomic.LoadInt32(&sm.itemCount)
	if itemCount == 0 {
		return false
	}

	deletedCount := atomic.LoadInt32(&sm.deletedCount)
	deletedRatio := float64(deletedCount) / float64(itemCount)

	lastShrink := sm.lastShrinkTime.Load().(time.Time)
	timeToShrink := time.Since(lastShrink) >= sm.config.MinShrinkInterval

	return deletedRatio >= sm.config.ShrinkRatio && timeToShrink
}

// shrink creates a new map and copies non-deleted items to it
func (sm *ShrinkableMap[K, V]) shrink() bool {
	// Prevent concurrent shrink operations
	if !sm.shrinking.CompareAndSwap(false, true) {
		return false
	}
	defer sm.shrinking.Store(false)

	startTime := time.Now()

	// Create snapshot of current data
	sm.mu.Lock()
	currentData := make(map[K]V, len(sm.data))
	for k, v := range sm.data {
		currentData[k] = v
	}
	sm.mu.Unlock()

	// Calculate new size
	currentLen := sm.Len()
	if currentLen == 0 {
		return false
	}

	newSize := int(float64(currentLen) * sm.config.CapacityGrowthFactor)
	if newSize < sm.config.InitialCapacity {
		newSize = sm.config.InitialCapacity
	}

	// Create and populate new map
	newMap := make(map[K]V, newSize)
	for k, v := range currentData {
		newMap[k] = v
	}

	// Update map with new data
	sm.mu.Lock()
	sm.data = newMap
	newCount := int32(len(newMap))
	atomic.StoreInt32(&sm.itemCount, newCount)
	atomic.StoreInt32(&sm.deletedCount, 0)
	sm.mu.Unlock()

	// Update metrics
	sm.metrics.mu.Lock()
	sm.metrics.totalShrinks++
	sm.metrics.lastShrinkDuration = time.Since(startTime)
	sm.metrics.mu.Unlock()

	sm.lastShrinkTime.Store(time.Now())

	return true
}

// TryShrink attempts to shrink the map if conditions are met
func (sm *ShrinkableMap[K, V]) TryShrink() bool {
	if sm.shouldShrink() {
		return sm.shrink()
	}
	return false
}

// ForceShrink immediately shrinks the map regardless of conditions
func (sm *ShrinkableMap[K, V]) ForceShrink() bool {
	return sm.shrink()
}

// shrinkLoop runs the periodic shrink check
func (sm *ShrinkableMap[K, V]) shrinkLoop() {
	ticker := time.NewTicker(sm.config.ShrinkInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.TryShrink()
	}
}
