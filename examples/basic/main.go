package main

import (
	"fmt"

	"github.com/jongyunha/shrinkmap"
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

	// Example 5: Error Handling
	errorHandlingExample()

	// Example 6: Snapshot
	snapshotExample()

	// Example 7: Resource Management
	resourceManagementExample()

	// Example 8: Comprehensive Example
	comprehensiveExample()

	// Example 9: Batch Operations
	batchExample()

	// Example 10: Iterator
	iteratorExample()

	// Example 11: Monitoring
	gradingSystemExample()
}

func basicExample() {
	// Create a map with default configuration
	sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())

	// Basic operations
	_ = sm.Set("one", 1)
	_ = sm.Set("two", 2)

	if val, exists := sm.Get("one"); exists {
		fmt.Printf("Basic - Value: %d\n", val)
	}

	sm.Delete("one")
	fmt.Printf("Basic - Current size: %d\n", sm.Len())
}
