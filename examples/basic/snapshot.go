package main

import (
	"fmt"

	"github.com/jongyunha/shrinkmap"
)

// snapshotExample demonstrates the use of the snapshot feature
func snapshotExample() {
	fmt.Println("\nSnapshot Example:")
	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
	defer sm.Stop()

	// Add some data
	testData := map[string]int{
		"apple":  1,
		"banana": 2,
		"cherry": 3,
		"date":   4,
	}

	for k, v := range testData {
		_ = sm.Set(k, v)
	}

	// Take a snapshot and process data safely
	snapshot := sm.Snapshot()

	// Process snapshot data without holding locks
	fmt.Println("Current map state:")
	for _, kv := range snapshot {
		fmt.Printf("- %s: %d\n", kv.Key, kv.Value)
	}

	// Demonstrate concurrent safety
	go func() {
		// Modify map while processing snapshot
		_ = sm.Set("new-key", 100)
		sm.Delete("apple")
	}()

	// Original snapshot remains unchanged
	fmt.Println("\nProcessing original snapshot again:")
	for _, kv := range snapshot {
		fmt.Printf("- %s: %d\n", kv.Key, kv.Value)
	}
}
