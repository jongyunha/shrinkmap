package shrinkmap

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ShrinkableMap provides a generic map structure with automatic shrinking capabilities
// Note: Each ShrinkableMap instance creates its own goroutine for auto-shrinking when AutoShrinkEnabled is true.
// The goroutine will continue to run until Stop() is called, even if there are no more references to the map.
// For transient use cases, ensure to call Stop() when the map is no longer needed to prevent goroutine leaks.
type ShrinkableMap[K comparable, V any] struct {
	mu             sync.RWMutex
	data           map[K]V
	itemCount      atomic.Int64
	deletedCount   atomic.Int64
	config         Config
	lastShrinkTime atomic.Value
	metrics        *Metrics
	shrinking      atomic.Bool
	cancel         context.CancelFunc
	stopped        atomic.Bool
}

// KeyValue represents a key-value pair for iteration purposes
type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// New creates a new ShrinkableMap with the given configuration
func New[K comparable, V any](config Config) *ShrinkableMap[K, V] {
	ctx, cancel := context.WithCancel(context.Background())
	sm := &ShrinkableMap[K, V]{
		data:    make(map[K]V, config.InitialCapacity),
		config:  config,
		metrics: &Metrics{},
		cancel:  cancel,
	}

	sm.lastShrinkTime.Store(time.Now())

	sm.itemCount.Store(0)
	sm.deletedCount.Store(0)

	if config.AutoShrinkEnabled {
		go sm.shrinkLoop(ctx)
	}
	return sm
}

// Stop terminates the auto-shrink goroutine if it's running
// This should be called when the map is no longer needed to prevent goroutine leaks
func (sm *ShrinkableMap[K, V]) Stop() {
	if sm.stopped.CompareAndSwap(false, true) {
		if sm.cancel != nil {
			sm.cancel()
		}
	}
}

// Snapshot returns a slice of key-value pairs representing the current state of the map
// Note: This operation requires a full lock of the map and may be expensive for large maps
func (sm *ShrinkableMap[K, V]) Snapshot() []KeyValue[K, V] {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]KeyValue[K, V], 0, len(sm.data))
	for k, v := range sm.data {
		result = append(result, KeyValue[K, V]{Key: k, Value: v})
	}
	return result
}

// Set stores a key-value pair in the map
func (sm *ShrinkableMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	_, exists := sm.data[key]
	sm.data[key] = value
	if !exists {
		sm.itemCount.Add(1)
		sm.updateMetrics(1)
	}
	needsShrink := sm.config.MaxMapSize > 0 && sm.itemCount.Load() >= int64(sm.config.MaxMapSize)
	sm.mu.Unlock()

	if needsShrink {
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
		sm.deletedCount.Add(1)
	}
	sm.mu.Unlock()

	if exists && sm.config.AutoShrinkEnabled {
		sm.TryShrink()
	}
	return exists
}

// Len returns the current number of items in the map
func (sm *ShrinkableMap[K, V]) Len() int64 {
	return sm.itemCount.Load() - sm.deletedCount.Load()
}

func (sm *ShrinkableMap[K, V]) updateMetrics(processedItems int64) {
	currentSize := sm.itemCount.Load()
	if currentSize > int64(atomic.LoadInt32(&sm.metrics.peakSize)) {
		sm.metrics.mu.Lock()
		sm.metrics.totalItemsProcessed += processedItems
		if currentSize > int64(sm.metrics.peakSize) {
			sm.metrics.peakSize = int32(currentSize)
		}
		sm.metrics.mu.Unlock()
	} else {
		atomic.AddInt64(&sm.metrics.totalItemsProcessed, processedItems)
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
		shrinkPanics:        sm.metrics.shrinkPanics,
		lastPanicTime:       sm.metrics.lastPanicTime,
		lastError:           sm.metrics.lastError,
		errorHistory:        sm.metrics.errorHistory,
		totalErrors:         sm.metrics.totalErrors,
	}
}

// shouldShrink determines if the map should be shrunk based on current conditions
func (sm *ShrinkableMap[K, V]) shouldShrink() bool {
	itemCount := sm.itemCount.Load()
	if itemCount == 0 {
		return false
	}

	deletedCount := sm.deletedCount.Load()
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

	// Calculate new size
	currentLen := sm.Len()
	if currentLen == 0 {
		return false
	}

	newSize := int(float64(currentLen) * sm.config.CapacityGrowthFactor)
	if newSize < sm.config.InitialCapacity {
		newSize = sm.config.InitialCapacity
	}

	sm.mu.Lock()
	// Create and populate new map
	newMap := make(map[K]V, newSize)
	for k, v := range sm.data {
		newMap[k] = v
	}
	// Update map with new data
	sm.data = newMap
	newCount := int64(len(newMap))
	sm.itemCount.Store(newCount)
	sm.deletedCount.Store(0)
	sm.mu.Unlock()

	sm.updateShrinkMetrics(startTime)
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

// shrinkLoop runs the periodic shrink check with panic recovery
func (sm *ShrinkableMap[K, V]) shrinkLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			sm.metrics.mu.Lock()
			sm.metrics.shrinkPanics++
			sm.metrics.lastPanicTime = time.Now()
			sm.metrics.mu.Unlock()
		}
	}()

	ticker := time.NewTicker(sm.config.ShrinkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sm.TryShrink()
		}
	}
}

func (sm *ShrinkableMap[K, V]) updateShrinkMetrics(startTime time.Time) {
	sm.metrics.mu.Lock()
	sm.metrics.totalShrinks++
	sm.metrics.lastShrinkDuration = time.Since(startTime)
	sm.metrics.mu.Unlock()
}
