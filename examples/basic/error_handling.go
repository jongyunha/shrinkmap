package main

import (
	"fmt"
	"time"

	"github.com/jongyunha/shrinkmap"
)

func errorHandlingExample() {
	fmt.Println("\nError Handling Example:")
	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
	defer sm.Stop()

	// Simulate some operations that might cause errors
	for i := 0; i < 100; i++ {
		_ = sm.Set(fmt.Sprintf("key-%d", i), i)
	}

	// Get error metrics
	metrics := sm.GetMetrics()

	// Check for any errors
	if metrics.TotalErrors() > 0 {
		fmt.Printf("Total errors encountered: %d\n", metrics.TotalErrors())

		// Get last error details
		if lastErr := metrics.LastError(); lastErr != nil {
			fmt.Printf("Last error: %v\n", lastErr.Error)
			fmt.Printf("Error time: %v\n", lastErr.Timestamp)
		}

		// Get error history
		fmt.Println("Recent error history:")
		for _, err := range metrics.ErrorHistory() {
			fmt.Printf("- [%v] %v\n", err.Timestamp.Format(time.RFC3339), err.Error)
		}
	}

	// Check panic statistics
	if metrics.TotalPanics() > 0 {
		fmt.Printf("Total panics recovered: %d\n", metrics.TotalPanics())
		fmt.Printf("Last panic time: %v\n", metrics.LastPanicTime())
	}
}
