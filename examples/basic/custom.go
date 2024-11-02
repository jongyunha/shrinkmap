package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jongyunha/shrinkmap"
)

func customExample() {
	// Create custom configuration
	config := shrinkmap.DefaultConfig().
		WithShrinkInterval(time.Minute).
		WithShrinkRatio(0.3).
		WithInitialCapacity(100).
		WithAutoShrinkEnabled(true).
		WithMinShrinkInterval(10 * time.Second).
		WithMaxMapSize(1000).
		WithCapacityGrowthFactor(1.5)

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create map with custom config
	sm := shrinkmap.New[string, int](config)

	// Add items
	sm.Set("one", 1)
	sm.Set("two", 2)
	sm.Set("three", 3)

	// Demonstrate get operation
	if val, exists := sm.Get("two"); exists {
		fmt.Printf("Custom - Found value: %d\n", val)
	}

	// Demonstrate delete and shrink
	sm.Delete("one")
	sm.ForceShrink() // Explicitly trigger shrink

	// Check metrics
	metrics := sm.GetMetrics()
	fmt.Printf("Custom - Total processed: %d, Total shrinks: %d\n",
		metrics.TotalItemsProcessed(), metrics.TotalShrinks())
}
