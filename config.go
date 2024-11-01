package shrinkmap

import (
	"fmt"
	"time"
)

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

// DefaultConfig returns the default configuration for ShrinkableMap
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

// Validate checks if the configuration is valid
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
