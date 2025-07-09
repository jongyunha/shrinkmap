package main

import (
	"fmt"
	"time"

	"github.com/jongyunha/shrinkmap"
)

func comprehensiveExample() {
	fmt.Println("\nComprehensive Example:")

	// Create map with custom configuration
	config := shrinkmap.DefaultConfig().
		WithAutoShrinkEnabled(true).
		WithShrinkInterval(time.Second).
		WithInitialCapacity(100)

	sm := shrinkmap.New[string, interface{}](config)
	defer sm.Stop() // Ensure cleanup

	// Simulate heavy usage
	go func() {
		for i := 0; i < 1000; i++ {
			_ = sm.Set(fmt.Sprintf("key-%d", i), i)
			if i%2 == 0 {
				sm.Delete(fmt.Sprintf("key-%d", i-1))
			}
			time.Sleep(time.Millisecond)
		}
	}()

	// Monitor and report (simulated monitoring routine)
	go func() {
		for i := 0; i < 5; i++ {
			metrics := sm.GetMetrics()
			snapshot := sm.Snapshot()

			fmt.Printf("\nStatus Report #%d:\n", i+1)
			fmt.Printf("- Current items: %d\n", len(snapshot))
			fmt.Printf("- Total operations: %d\n", metrics.TotalItemsProcessed())
			fmt.Printf("- Total errors: %d\n", metrics.TotalErrors())
			fmt.Printf("- Total panics: %d\n", metrics.TotalPanics())

			time.Sleep(time.Second)
		}
	}()

	// Wait for demonstration to complete
	time.Sleep(5 * time.Second)
}
