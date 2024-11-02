package main

import (
	"fmt"

	"github.com/jongyunha/shrinkmap"
)

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
