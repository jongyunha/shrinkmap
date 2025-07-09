// Package shrinkmap provides a thread-safe, generic map implementation with automatic memory management.
//
// ShrinkableMap automatically shrinks its internal storage when items are deleted,
// addressing the common issue where Go's built-in maps don't release memory after deleting elements.
//
// Key features:
//   - Thread-safe operations with minimal locking overhead
//   - Automatic memory shrinking with configurable policies
//   - Generic type support for compile-time type safety
//   - Comprehensive error handling with structured error types
//   - Performance metrics and monitoring capabilities
//   - Batch operations for improved performance
//   - Safe iteration with snapshot support
//
// Basic usage:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	defer sm.Stop() // Always call Stop() to prevent goroutine leaks
//
//	// Basic operations
//	sm.Set("key1", 42)
//	if value, exists := sm.Get("key1"); exists {
//		fmt.Printf("Value: %d\n", value)
//	}
//	sm.Delete("key1")
//
// For production use, always call Stop() when the map is no longer needed
// to prevent goroutine leaks from the automatic shrinking background process.
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
	// 64-bit aligned fields (should be first on 32-bit architectures)
	itemCount      atomic.Int64
	deletedCount   atomic.Int64
	
	// Pointers and interfaces (8 bytes on 64-bit, 4 bytes on 32-bit)
	data           map[K]V
	metrics        *Metrics
	cancel         context.CancelFunc
	lastShrinkTime atomic.Value
	
	// Mutex (24 bytes on 64-bit)
	mu             sync.RWMutex
	
	// Structs (depends on size)
	config         Config
	
	// Atomic bools (1 byte each but aligned)
	shrinking      atomic.Bool
	stopped        atomic.Bool
}

// KeyValue represents a key-value pair for iteration purposes
type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// New creates a new ShrinkableMap with the given configuration.
// The returned map is ready for concurrent use and will automatically
// start background shrinking if AutoShrinkEnabled is true in the config.
//
// Important: Always call Stop() when the map is no longer needed to
// prevent goroutine leaks from the automatic shrinking process.
//
// Example:
//
//	config := shrinkmap.DefaultConfig()
//	config.ShrinkRatio = 0.5 // Shrink when 50% of items are deleted
//	sm := shrinkmap.New[string, int](config)
//	defer sm.Stop()
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

// Stop terminates the auto-shrink goroutine if it's running.
// This should be called when the map is no longer needed to prevent goroutine leaks.
// After calling Stop(), the map remains functional but automatic shrinking is disabled.
// It's safe to call Stop() multiple times.
//
// Example:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	defer sm.Stop() // Ensure cleanup
//	
//	// Use the map...
//	sm.Set("key", 42)
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
func (sm *ShrinkableMap[K, V]) Set(key K, value V) error {
	if sm.stopped.Load() {
		return ErrMapStopped
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check capacity before adding new item
	if sm.config.MaxMapSize > 0 && !sm.keyExists(key) && sm.itemCount.Load() >= int64(sm.config.MaxMapSize) {
		return ErrCapacityExceeded.WithDetails("currentSize", sm.itemCount.Load()).WithDetails("maxSize", sm.config.MaxMapSize)
	}

	_, exists := sm.data[key]
	sm.data[key] = value
	if !exists {
		sm.itemCount.Add(1)
		sm.updateMetrics(1)
	}

	needsShrink := sm.config.MaxMapSize > 0 && sm.itemCount.Load() >= int64(sm.config.MaxMapSize)
	if needsShrink {
		go sm.TryShrink()
	}

	return nil
}

// keyExists checks if a key exists without acquiring additional locks
func (sm *ShrinkableMap[K, V]) keyExists(key K) bool {
	_, exists := sm.data[key]
	return exists
}

// Get retrieves the value associated with the given key.
// Returns the value and a boolean indicating whether the key exists.
// This operation is safe for concurrent use.
//
// Example:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	sm.Set("key", 42)
//	if value, exists := sm.Get("key"); exists {
//		fmt.Printf("Value: %d\n", value)
//	}
func (sm *ShrinkableMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	value, exists := sm.data[key]
	sm.mu.RUnlock()
	return value, exists
}

// Delete removes the entry for the given key.
// Returns true if the key existed and was deleted, false otherwise.
// This operation is safe for concurrent use and may trigger automatic shrinking.
//
// Example:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	sm.Set("key", 42)
//	if deleted := sm.Delete("key"); deleted {
//		fmt.Println("Key was deleted")
//	}
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

// Len returns the current number of items in the map.
// This operation is atomic and safe for concurrent use.
// The returned value reflects the actual number of accessible items,
// accounting for deletions that haven't been cleaned up yet.
//
// Example:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	sm.Set("key1", 1)
//	sm.Set("key2", 2)
//	fmt.Printf("Map contains %d items\n", sm.Len()) // Output: Map contains 2 items
func (sm *ShrinkableMap[K, V]) Len() int64 {
	return sm.itemCount.Load() - sm.deletedCount.Load()
}

// Contains checks if the key exists in the map.
// This is a convenience method that returns only the existence boolean.
// For better performance when you also need the value, use Get instead.
//
// Example:
//
//	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
//	sm.Set("key", 42)
//	if sm.Contains("key") {
//		fmt.Println("Key exists")
//	}
func (sm *ShrinkableMap[K, V]) Contains(key K) bool {
	sm.mu.RLock()
	_, exists := sm.data[key]
	sm.mu.RUnlock()
	return exists
}

// SetIf sets the value for the key only if the condition is true
func (sm *ShrinkableMap[K, V]) SetIf(key K, value V, condition func(oldValue V, exists bool) bool) (bool, error) {
	if sm.stopped.Load() {
		return false, ErrMapStopped
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldValue, exists := sm.data[key]
	if condition(oldValue, exists) {
		// Check capacity before adding new item
		if sm.config.MaxMapSize > 0 && !exists && sm.itemCount.Load() >= int64(sm.config.MaxMapSize) {
			return false, ErrCapacityExceeded.WithDetails("currentSize", sm.itemCount.Load()).WithDetails("maxSize", sm.config.MaxMapSize)
		}

		sm.data[key] = value
		if !exists {
			sm.itemCount.Add(1)
			sm.updateMetrics(1)
		}
		return true, nil
	}
	return false, nil
}

// GetOrSet returns the existing value for the key, or sets and returns the provided value if key doesn't exist
func (sm *ShrinkableMap[K, V]) GetOrSet(key K, value V) (V, bool, error) {
	if sm.stopped.Load() {
		var zero V
		return zero, false, ErrMapStopped
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if existingValue, exists := sm.data[key]; exists {
		return existingValue, true, nil
	}

	// Check capacity before adding new item
	if sm.config.MaxMapSize > 0 && sm.itemCount.Load() >= int64(sm.config.MaxMapSize) {
		var zero V
		return zero, false, ErrCapacityExceeded.WithDetails("currentSize", sm.itemCount.Load()).WithDetails("maxSize", sm.config.MaxMapSize)
	}

	sm.data[key] = value
	sm.itemCount.Add(1)
	sm.updateMetrics(1)
	return value, false, nil
}

// SetIfAbsent sets the value for the key only if the key doesn't exist
func (sm *ShrinkableMap[K, V]) SetIfAbsent(key K, value V) (bool, error) {
	return sm.SetIf(key, value, func(oldValue V, exists bool) bool {
		return !exists
	})
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
