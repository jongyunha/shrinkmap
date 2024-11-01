package main

import (
	"fmt"
	"github.com/jongyunha/shrinkmap"
	"log"
	"time"
)

func main() {
	// Example 1: Using Default Configuration
	basicExample()

	// Example 2: Using Custom Configuration
	customExample()

	// Example 3: Using Builder Pattern Configuration
	builderExample()

	// Example 4: Monitoring Usage
	monitoringExample()
}

func basicExample() {
	// Create a map with default configuration
	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())

	// Basic operations
	sm.Set("one", 1)
	sm.Set("two", 2)

	if val, exists := sm.Get("one"); exists {
		fmt.Printf("Basic - Value: %d\n", val)
	}

	sm.Delete("one")
	fmt.Printf("Basic - Current size: %d\n", sm.Len())
}

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
		metrics.TotalItemsProcessed, metrics.TotalShrinks)
}

func builderExample() {
	// Create configuration using builder pattern
	config := shrinkmap.DefaultConfig().
		WithShrinkInterval(30 * time.Second).
		WithShrinkRatio(0.2)

	sm := shrinkmap.New[int, string](config)

	// Demonstrate typical usage pattern
	for i := 0; i < 100; i++ {
		sm.Set(i, fmt.Sprintf("value-%d", i))
	}

	// Delete some items to trigger shrinking
	for i := 0; i < 30; i++ {
		sm.Delete(i)
	}

	// Try to shrink
	if shrunk := sm.TryShrink(); shrunk {
		fmt.Println("Builder - Map was shrunk")
	}
}

func monitoringExample() {
	sm := shrinkmap.New[string, interface{}](shrinkmap.DefaultConfig())

	// Add various types of data
	sm.Set("string", "hello")
	sm.Set("number", 42)
	sm.Set("bool", true)

	// Get metrics
	metrics := sm.GetMetrics()

	// Print detailed metrics
	fmt.Printf("Monitoring Results:\n")
	fmt.Printf("- Total processed items: %d\n", metrics.TotalItemsProcessed())
	fmt.Printf("- Total shrink operations: %d\n", metrics.TotalShrinks())
	fmt.Printf("- Peak size: %d\n", metrics.PeakSize())
	if metrics.LastShrinkDuration() > 0 {
		fmt.Printf("- Last shrink duration: %v\n", metrics.LastShrinkDuration())
	}
}
