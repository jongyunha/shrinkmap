package shrinkmap

import "time"

// Builder provides a fluent interface for creating and configuring ShrinkableMap instances.
// It allows method chaining for a more readable and convenient API.
type Builder[K comparable, V any] struct {
	config Config
}

// NewBuilder creates a new Builder with default configuration.
// This is the starting point for creating a ShrinkableMap with fluent syntax.
//
// Example:
//
//	sm := shrinkmap.NewBuilder[string, int]().
//		WithShrinkRatio(0.5).
//		WithInitialCapacity(100).
//		Build()
//	defer sm.Stop()
func NewBuilder[K comparable, V any]() *Builder[K, V] {
	return &Builder[K, V]{
		config: DefaultConfig(),
	}
}

// WithShrinkRatio sets the shrink ratio and returns the builder for chaining.
// The shrink ratio determines what percentage of deleted items triggers shrinking.
func (b *Builder[K, V]) WithShrinkRatio(ratio float64) *Builder[K, V] {
	b.config.ShrinkRatio = ratio
	return b
}

// WithInitialCapacity sets the initial capacity and returns the builder for chaining.
// This is the initial capacity of the underlying map.
func (b *Builder[K, V]) WithInitialCapacity(capacity int) *Builder[K, V] {
	b.config.InitialCapacity = capacity
	return b
}

// WithShrinkInterval sets the shrink interval and returns the builder for chaining.
// This controls how often the map checks for shrinking opportunities.
func (b *Builder[K, V]) WithShrinkInterval(interval time.Duration) *Builder[K, V] {
	b.config.ShrinkInterval = interval
	return b
}

// WithMinShrinkInterval sets the minimum shrink interval and returns the builder for chaining.
// This prevents shrinking too frequently even if conditions are met.
func (b *Builder[K, V]) WithMinShrinkInterval(interval time.Duration) *Builder[K, V] {
	b.config.MinShrinkInterval = interval
	return b
}

// WithMaxMapSize sets the maximum map size and returns the builder for chaining.
// Set to 0 for unlimited size.
func (b *Builder[K, V]) WithMaxMapSize(size int) *Builder[K, V] {
	b.config.MaxMapSize = size
	return b
}

// WithCapacityGrowthFactor sets the capacity growth factor and returns the builder for chaining.
// This controls how much extra space is allocated when shrinking.
func (b *Builder[K, V]) WithCapacityGrowthFactor(factor float64) *Builder[K, V] {
	b.config.CapacityGrowthFactor = factor
	return b
}

// WithAutoShrink enables or disables automatic shrinking and returns the builder for chaining.
func (b *Builder[K, V]) WithAutoShrink(enabled bool) *Builder[K, V] {
	b.config.AutoShrinkEnabled = enabled
	return b
}

// Build creates the ShrinkableMap with the configured settings.
// This is the final method in the builder chain that returns the actual map.
// The returned map is ready for use and will start background shrinking if enabled.
//
// Important: Always call Stop() on the returned map when done to prevent goroutine leaks.
func (b *Builder[K, V]) Build() *ShrinkableMap[K, V] {
	return New[K, V](b.config)
}

// BuildWithValidation creates the ShrinkableMap with validation of the configuration.
// Returns an error if the configuration is invalid.
func (b *Builder[K, V]) BuildWithValidation() (*ShrinkableMap[K, V], error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return New[K, V](b.config), nil
}

// MapBuilder provides additional fluent methods for map operations after creation.
type MapBuilder[K comparable, V any] struct {
	sm *ShrinkableMap[K, V]
}

// NewMapBuilder creates a MapBuilder from an existing ShrinkableMap.
// This allows for fluent operations on the map.
func NewMapBuilder[K comparable, V any](sm *ShrinkableMap[K, V]) *MapBuilder[K, V] {
	return &MapBuilder[K, V]{sm: sm}
}

// Set stores a key-value pair and returns the builder for chaining.
// If an error occurs, it can be retrieved using the LastError method.
func (mb *MapBuilder[K, V]) Set(key K, value V) *MapBuilder[K, V] {
	_ = mb.sm.Set(key, value)
	return mb
}

// Delete removes a key and returns the builder for chaining.
func (mb *MapBuilder[K, V]) Delete(key K) *MapBuilder[K, V] {
	mb.sm.Delete(key)
	return mb
}

// SetIfAbsent sets a key-value pair only if the key doesn't exist and returns the builder for chaining.
func (mb *MapBuilder[K, V]) SetIfAbsent(key K, value V) *MapBuilder[K, V] {
	_, _ = mb.sm.SetIfAbsent(key, value)
	return mb
}

// Map returns the underlying ShrinkableMap.
func (mb *MapBuilder[K, V]) Map() *ShrinkableMap[K, V] {
	return mb.sm
}

// Done is a convenience method that returns the underlying map.
// This is useful at the end of a builder chain.
func (mb *MapBuilder[K, V]) Done() *ShrinkableMap[K, V] {
	return mb.sm
}
