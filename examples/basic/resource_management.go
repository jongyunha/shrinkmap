package main

import (
	"fmt"
	"time"

	"github.com/jongyunha/shrinkmap"
)

// resourceManagementExample demonstrates proper resource cleanup and management
func resourceManagementExample() {
	fmt.Println("\nResource Management Example:")

	// Create map with auto-shrink enabled
	config := shrinkmap.DefaultConfig().
		WithAutoShrinkEnabled(true).
		WithShrinkInterval(time.Second)

	sm := shrinkmap.New[string, int](config)

	// Proper cleanup with defer
	defer func() {
		sm.Stop()
		fmt.Println("Map resources cleaned up")
	}()

	// Simulate some work
	fmt.Println("Performing operations...")
	for i := 0; i < 50; i++ {
		_ = sm.Set(fmt.Sprintf("key-%d", i), i)
	}

	// Delete some items
	for i := 0; i < 20; i++ {
		sm.Delete(fmt.Sprintf("key-%d", i))
	}

	// Get current metrics before cleanup
	metrics := sm.GetMetrics()
	fmt.Printf("Final metrics before cleanup:\n")
	fmt.Printf("- Total items processed: %d\n", metrics.TotalItemsProcessed())
	fmt.Printf("- Current size: %d\n", sm.Len())
	fmt.Printf("- Total shrinks: %d\n", metrics.TotalShrinks())

	// Take final snapshot
	snapshot := sm.Snapshot()
	fmt.Printf("- Items remaining: %d\n", len(snapshot))
}
