package main

import (
	"fmt"
	"time"

	"github.com/jongyunha/shrinkmap"
)

func builderExample() {
	// Create configuration using builder pattern
	config := shrinkmap.DefaultConfig().
		WithShrinkInterval(30 * time.Second).
		WithShrinkRatio(0.2)

	sm := shrinkmap.New[int, string](config)

	// Demonstrate typical usage pattern
	for i := 0; i < 100; i++ {
		_ = sm.Set(i, fmt.Sprintf("value-%d", i))
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
