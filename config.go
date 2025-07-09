package shrinkmap

import (
	"fmt"
	"time"
)

// Config defines the configuration options for ShrinkableMap
type Config struct {
	// Duration values (8 bytes each)
	ShrinkInterval    time.Duration // How often to check if the map needs shrinking
	MinShrinkInterval time.Duration // Minimum time between shrinks

	// Float64 values (8 bytes each)
	ShrinkRatio          float64 // Ratio of deleted items that triggers shrinking (0.0 to 1.0)
	CapacityGrowthFactor float64 // Extra capacity factor when creating new map (e.g., 1.2 for 20% extra space)

	// Int values (8 bytes on 64-bit)
	InitialCapacity int // Initial capacity of the map
	MaxMapSize      int // Maximum map size before forcing a shrink

	// Bool values (1 byte each)
	AutoShrinkEnabled bool // Enable/disable automatic shrinking
}

// DefaultConfig returns the default configuration for ShrinkableMap.
// These settings are optimized for general use cases with reasonable
// performance and memory management characteristics.
//
// Default values:
//   - ShrinkInterval: 5 minutes (how often to check for shrinking)
//   - ShrinkRatio: 0.25 (shrink when 25% of items are deleted)
//   - InitialCapacity: 16 (starting capacity)
//   - AutoShrinkEnabled: true (automatic shrinking is enabled)
//   - MinShrinkInterval: 30 seconds (minimum time between shrinks)
//   - MaxMapSize: 1,000,000 (maximum items before forced shrink)
//   - CapacityGrowthFactor: 1.2 (20% extra space when shrinking)
//
// Example:
//
//	config := shrinkmap.DefaultConfig()
//	config.ShrinkRatio = 0.5 // Customize shrink ratio
//	sm := shrinkmap.New[string, int](config)
func DefaultConfig() Config {
	return Config{
		// Check for shrinking every 5 minutes
		ShrinkInterval: 5 * time.Minute,

		// Shrink when 25% of items are deleted
		ShrinkRatio: 0.25,

		// Start with a reasonable initial capacity
		InitialCapacity: 16,

		// Enable automatic shrinking by default
		AutoShrinkEnabled: true,

		// Wait at least 30 seconds between shrinks
		MinShrinkInterval: 30 * time.Second,

		// Set a reasonable maximum size (1 million items)
		// Use 0 for unlimited
		MaxMapSize: 1_000_000,

		// Allocate 20% extra space when shrinking
		CapacityGrowthFactor: 1.2,
	}
}

// WithShrinkInterval sets the shrink interval and returns the modified config
func (c Config) WithShrinkInterval(d time.Duration) Config {
	c.ShrinkInterval = d
	return c
}

// WithShrinkRatio sets the shrink ratio and returns the modified config
func (c Config) WithShrinkRatio(ratio float64) Config {
	c.ShrinkRatio = ratio
	return c
}

// WithInitialCapacity sets the initial capacity and returns the modified config
func (c Config) WithInitialCapacity(capacity int) Config {
	c.InitialCapacity = capacity
	return c
}

// WithAutoShrinkEnabled sets auto shrinking and returns the modified config
func (c Config) WithAutoShrinkEnabled(enabled bool) Config {
	c.AutoShrinkEnabled = enabled
	return c
}

// WithMinShrinkInterval sets the minimum shrink interval and returns the modified config
func (c Config) WithMinShrinkInterval(d time.Duration) Config {
	c.MinShrinkInterval = d
	return c
}

// WithMaxMapSize sets the maximum map size and returns the modified config
func (c Config) WithMaxMapSize(size int) Config {
	c.MaxMapSize = size
	return c
}

// WithCapacityGrowthFactor sets the capacity growth factor and returns the modified config
func (c Config) WithCapacityGrowthFactor(factor float64) Config {
	c.CapacityGrowthFactor = factor
	return c
}

// Validate checks if the configuration is valid and returns an error if not.
// This method should be called before creating a new ShrinkableMap to ensure
// the configuration parameters are within acceptable ranges.
//
// Validation rules:
//   - ShrinkInterval must be positive
//   - ShrinkRatio must be between 0 and 1 (exclusive)
//   - InitialCapacity must be non-negative
//   - MinShrinkInterval must be positive
//   - MaxMapSize must be non-negative (0 means unlimited)
//   - CapacityGrowthFactor must be greater than 1
//
// Example:
//
//	config := shrinkmap.DefaultConfig()
//	config.ShrinkRatio = 0.5
//	if err := config.Validate(); err != nil {
//		log.Fatal("Invalid configuration:", err)
//	}
func (c Config) Validate() error {
	if c.ShrinkInterval <= 0 {
		return fmt.Errorf("shrink interval must be positive")
	}
	if c.ShrinkRatio <= 0 || c.ShrinkRatio >= 1 {
		return fmt.Errorf("shrink ratio must be between 0 and 1")
	}
	if c.InitialCapacity < 0 {
		return fmt.Errorf("initial capacity must be non-negative")
	}
	if c.MinShrinkInterval <= 0 {
		return fmt.Errorf("minimum shrink interval must be positive")
	}
	if c.MaxMapSize < 0 {
		return fmt.Errorf("maximum map size must be non-negative")
	}
	if c.CapacityGrowthFactor <= 1 {
		return fmt.Errorf("capacity growth factor must be greater than 1")
	}
	return nil
}
