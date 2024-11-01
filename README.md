# ShrinkableMap

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

ShrinkableMap is a high-performance, generic, thread-safe map implementation for Go that automatically manages memory by shrinking its internal storage when items are deleted. It provides a solution to the common issue where Go's built-in maps don't release memory after deleting elements.

## Features

- 🚀 Generic type support for type-safe operations
- 🔒 Thread-safe implementation with atomic operations
- 📉 Automatic memory shrinking with configurable policies
- ⚙️ Advanced concurrent shrinking behavior
- 📊 Thread-safe performance metrics
- 🎯 Zero external dependencies
- 💪 Production-ready with comprehensive tests
- 🛡️ Race condition free

## Installation

```bash
go get github.com/jongyunha/shrinkmap
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/jongyunha/shrinkmap"
)

func main() {
    // Create a new map with string keys and int values
    sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
    
    // Set values
    sm.Set("one", 1)
    sm.Set("two", 2)
    
    // Get value
    if value, exists := sm.Get("one"); exists {
        fmt.Printf("Value: %d\n", value)
    }
    
    // Delete value
    sm.Delete("one")
    
    // Get current size
    size := sm.Len()
    fmt.Printf("Map size: %d\n", size)
    
    // Force shrink
    sm.ForceShrink()
    
    // Get metrics
    metrics := sm.GetMetrics()
    fmt.Printf("Total operations: %d\n", metrics.TotalItemsProcessed)
}
```

## Configuration

```go
// Create default configuration
config := shrinkmap.DefaultConfig()

// Or customize configuration
config := shrinkmap.Config{
    ShrinkInterval:        5 * time.Minute,  // How often to check for shrinking
    ShrinkRatio:          0.25,             // Ratio of deleted items that triggers shrinking
    InitialCapacity:      16,               // Initial map capacity
    AutoShrinkEnabled:    true,             // Enable automatic shrinking
    MinShrinkInterval:    30 * time.Second, // Minimum time between shrinks
    MaxMapSize:           1000000,          // Maximum map size before forcing shrink
    CapacityGrowthFactor: 1.2,             // Growth factor for new map allocation
}

// Create map with custom config
sm := shrinkmap.New[string, int](config)

// Use builder pattern for configuration
config := shrinkmap.DefaultConfig().
    WithShrinkInterval(time.Minute).
    WithShrinkRatio(0.3).
    WithInitialCapacity(1000).
    WithAutoShrinkEnabled(true)
```

## Features in Detail

### Atomic Operations

All operations are implemented using atomic operations where appropriate, ensuring thread safety without compromising performance:

```go
// Safely modify map in multiple goroutines
for i := 0; i < 100; i++ {
    go func(val int) {
        sm.Set(fmt.Sprintf("key%d", val), val)
    }(i)
}
```

### Thread-Safe Metrics

Monitor map performance with built-in thread-safe metrics:

```go
metrics := sm.GetMetrics()
fmt.Printf("Total shrinks: %d\n", metrics.TotalShrinks)
fmt.Printf("Last shrink duration: %v\n", metrics.LastShrinkDuration)
fmt.Printf("Total items processed: %d\n", metrics.TotalItemsProcessed)
fmt.Printf("Peak size: %d\n", metrics.PeakSize)
```

### Shrinking Control

Fine-grained control over shrinking behavior:

```go
// Force immediate shrink
success := sm.ForceShrink()

// Try to shrink if conditions are met
shrunk := sm.TryShrink()

// Configure shrinking behavior
config := shrinkmap.DefaultConfig().
    WithShrinkRatio(0.3).            // More aggressive shrinking
    WithMinShrinkInterval(time.Minute)  // More frequent shrinks
```

## Performance

Benchmark results (Go 1.22, AMD64):

```
BenchmarkShrinkableMap/Set/Parallel-8          10000000    115 ns/op     0 allocs/op
BenchmarkShrinkableMap/Get/Parallel-8          20000000    60.1 ns/op    0 allocs/op
BenchmarkShrinkableMap/Delete/Parallel-8       15000000    102 ns/op     0 allocs/op
BenchmarkShrinkableMap/Mixed/Parallel-8        10000000    112 ns/op     0 allocs/op
```

## Best Practices

1. Configure shrinking based on your usage pattern:
```go
// High-churn scenario
config := shrinkmap.DefaultConfig().
    WithShrinkInterval(time.Minute).
    WithShrinkRatio(0.2)

// Memory-sensitive scenario
config := shrinkmap.DefaultConfig().
    WithCapacityGrowthFactor(1.1).
    WithMaxMapSize(100000)
```

2. Use appropriate initial capacity:
```go
// When you know approximate size
config := shrinkmap.DefaultConfig().
    WithInitialCapacity(expectedSize)
```

3. Validate configuration:
```go
config := shrinkmap.DefaultConfig().
    WithShrinkRatio(0.3)

if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

## Thread Safety Guarantees

- All map operations are atomic and thread-safe
- Metrics collection is non-blocking and thread-safe
- Shrinking operations are coordinated to prevent conflicts
- Safe concurrent access from multiple goroutines

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Go's built-in map implementation
- Built with lessons learned from production systems
- Special thanks to all contributors

## Version History

- 0.0.1 (ing...)
  - Initial release
  - Thread-safe implementation with atomic operations
  - Generic type support
  - Automatic shrinking with configurable policies
  - Comprehensive benchmark suite
  - Race condition free guarantee

---
Made with ❤️ by [Jongyun Ha](https://github.com/jongyunha)