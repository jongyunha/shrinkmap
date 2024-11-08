package main

import (
	"fmt"

	"github.com/jongyunha/shrinkmap"
)

// BatchExample demonstrates various batch operations
func batchExample() {
	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
	defer sm.Stop()

	// Prepare batch operations
	batch := shrinkmap.BatchOperations[string, int]{
		Operations: []shrinkmap.BatchOperation[string, int]{
			{Type: shrinkmap.BatchSet, Key: "user1_score", Value: 100},
			{Type: shrinkmap.BatchSet, Key: "user2_score", Value: 85},
			{Type: shrinkmap.BatchSet, Key: "user3_score", Value: 95},
			{Type: shrinkmap.BatchSet, Key: "user4_score", Value: 75},
		},
	}

	// Apply batch operations atomically
	if err := sm.ApplyBatch(batch); err != nil {
		fmt.Printf("Batch operation failed: %v\n", err)
		return
	}

	// Prepare update batch
	updateBatch := shrinkmap.BatchOperations[string, int]{
		Operations: []shrinkmap.BatchOperation[string, int]{
			{Type: shrinkmap.BatchSet, Key: "user1_score", Value: 95}, // Update score
			{Type: shrinkmap.BatchDelete, Key: "user4_score"},         // Remove user
			{Type: shrinkmap.BatchSet, Key: "user5_score", Value: 88}, // Add new user
		},
	}

	// Apply update batch
	if err := sm.ApplyBatch(updateBatch); err != nil {
		fmt.Printf("Update batch operation failed: %v\n", err)
		return
	}

	// Print final state using iterator
	iter := sm.NewIterator()
	fmt.Println("\nFinal Scores:")
	for iter.Next() {
		key, value := iter.Get()
		fmt.Printf("%s: %d\n", key, value)
	}
}
