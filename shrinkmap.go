package shrinkmap

import (
	"sync"
	"time"
)

// ShrinkableMap provides a generic map structure with automatic shrinking capabilities
type ShrinkableMap[K comparable, V any] struct {
	data           map[K]V
	mu             sync.RWMutex
	itemCount      int
	deletedCount   int
	config         Config
	lastShrinkTime time.Time
	metrics        Metrics
}

// Config defines the configuration options for ShrinkableMap
type Config struct {
	// How often to check if the map needs shrinking
	ShrinkInterval time.Duration
	// Ratio of deleted items that triggers shrinking (0.0 to 1.0)
	ShrinkRatio float64
	// Initial capacity of the map
	InitialCapacity int
	// Enable/disable automatic shrinking
	AutoShrinkEnabled bool
	// Minimum time between shrinks
	MinShrinkInterval time.Duration
	// Maximum map size before forcing a shrink
	MaxMapSize int
	// Extra capacity factor when creating new map (e.g., 1.2 for 20% extra space)
	CapacityGrowthFactor float64
}

// Metrics tracks performance metrics of the map
type Metrics struct {
	TotalShrinks        int
	LastShrinkDuration  time.Duration
	TotalItemsProcessed int64
	PeakSize            int
}

// DefaultConfig returns the default configuration for ShrinkableMap
func DefaultConfig() Config {
	return Config{
		ShrinkInterval:       5 * time.Minute,
		ShrinkRatio:          0.25,
		InitialCapacity:      16,
		AutoShrinkEnabled:    true,
		MinShrinkInterval:    30 * time.Second,
		MaxMapSize:           1000000,
		CapacityGrowthFactor: 1.2,
	}
}

// New creates a new ShrinkableMap with the given configuration
func New[K comparable, V any](config Config) *ShrinkableMap[K, V] {
	sm := &ShrinkableMap[K, V]{
		data:           make(map[K]V, config.InitialCapacity),
		config:         config,
		lastShrinkTime: time.Now(),
	}

	if config.AutoShrinkEnabled {
		go sm.shrinkLoop()
	}
	return sm
}

// Set stores a key-value pair in the map
func (sm *ShrinkableMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.data[key]; !exists {
		sm.itemCount++
		sm.metrics.TotalItemsProcessed++
		if sm.itemCount > sm.metrics.PeakSize {
			sm.metrics.PeakSize = sm.itemCount
		}
	}
	sm.data[key] = value

	// Force shrink if max size is exceeded
	if sm.config.MaxMapSize > 0 && sm.itemCount >= sm.config.MaxMapSize {
		sm.shrink()
	}
}

// Get retrieves the value associated with the given key
func (sm *ShrinkableMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	value, exists := sm.data[key]
	return value, exists
}

// Delete removes the entry for the given key
func (sm *ShrinkableMap[K, V]) Delete(key K) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.data[key]; exists {
		delete(sm.data, key)
		sm.deletedCount++
		if sm.config.AutoShrinkEnabled {
			sm.checkAndShrink()
		}
		return true
	}
	return false
}

// Len returns the current number of items in the map
func (sm *ShrinkableMap[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.itemCount - sm.deletedCount
}

// ForceShrink immediately shrinks the map regardless of conditions
func (sm *ShrinkableMap[K, V]) ForceShrink() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.shrink()
}

// TryShrink attempts to shrink the map if conditions are met
func (sm *ShrinkableMap[K, V]) TryShrink() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.shouldShrink() {
		return sm.shrink()
	}
	return false
}

// GetMetrics returns the current performance metrics
func (sm *ShrinkableMap[K, V]) GetMetrics() Metrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.metrics
}

// UpdateConfig updates the map's configuration at runtime
func (sm *ShrinkableMap[K, V]) UpdateConfig(newConfig Config) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.config = newConfig
}

// Range iterates over the map and calls the given function for each key-value pair
func (sm *ShrinkableMap[K, V]) Range(fn func(key K, value V) bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for k, v := range sm.data {
		if !fn(k, v) {
			break
		}
	}
}

// shrinkLoop runs the periodic shrink check
func (sm *ShrinkableMap[K, V]) shrinkLoop() {
	ticker := time.NewTicker(sm.config.ShrinkInterval)
	for range ticker.C {
		sm.mu.Lock()
		sm.checkAndShrink()
		sm.mu.Unlock()
	}
}

// checkAndShrink checks if shrinking is needed and performs it if necessary
func (sm *ShrinkableMap[K, V]) checkAndShrink() {
	if sm.shouldShrink() {
		sm.shrink()
	}
}

// shouldShrink determines if the map should be shrunk based on current conditions
func (sm *ShrinkableMap[K, V]) shouldShrink() bool {
	if sm.itemCount == 0 {
		return false
	}

	deletedRatio := float64(sm.deletedCount) / float64(sm.itemCount)
	timeToShrink := time.Since(sm.lastShrinkTime) >= sm.config.MinShrinkInterval

	return deletedRatio >= sm.config.ShrinkRatio && timeToShrink
}

// shrink creates a new map and copies non-deleted items to it
func (sm *ShrinkableMap[K, V]) shrink() bool {
	if sm.itemCount == 0 {
		return false
	}

	startTime := time.Now()
	newSize := int(float64(sm.Len()) * sm.config.CapacityGrowthFactor)
	if newSize < sm.config.InitialCapacity {
		newSize = sm.config.InitialCapacity
	}

	newMap := make(map[K]V, newSize)

	for k, v := range sm.data {
		newMap[k] = v
	}

	sm.data = newMap
	sm.itemCount = len(newMap)
	sm.deletedCount = 0
	sm.lastShrinkTime = time.Now()

	sm.metrics.TotalShrinks++
	sm.metrics.LastShrinkDuration = time.Since(startTime)

	return true
}
