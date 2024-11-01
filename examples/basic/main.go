package main

import (
	"fmt"
	"github.com/jongyunha/shrinkmap"
	"time"
)

func main() {
	// Custom configuration
	config := shrinkmap.Config{
		ShrinkInterval:       time.Minute,
		ShrinkRatio:          0.3,
		InitialCapacity:      100,
		AutoShrinkEnabled:    true,
		MinShrinkInterval:    10 * time.Second,
		MaxMapSize:           1000,
		CapacityGrowthFactor: 1.5,
	}

	// Create a new map
	sm := shrinkmap.New[string, int](config)

	// Add some items
	sm.Set("one", 1)
	sm.Set("two", 2)
	sm.Set("three", 3)

	// Get and print values
	if val, exists := sm.Get("two"); exists {
		fmt.Printf("Value: %d\n", val)
	}

	// Delete an item
	sm.Delete("one")

	// Print metrics
	metrics := sm.GetMetrics()
	fmt.Printf("Total processed: %d\n", metrics.TotalItemsProcessed)
}
