# ShrinkableMap

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

ShrinkableMap is a high-performance, generic, thread-safe map implementation for Go that automatically manages memory by shrinking its internal storage when items are deleted. It provides a solution to the common issue where Go's built-in maps don't release memory after deleting elements.

## Features

- 🚀 Generic type support for type-safe operations
- 🔒 Thread-safe implementation
- 📉 Automatic memory shrinking
- ⚙️ Configurable shrinking behavior
- 📊 Built-in performance metrics
- 🎯 Zero external dependencies
- 💪 Production-ready with comprehensive tests

## Installation

```bash
go get github.com/jongyunha/shrinkmap
```

## Quick Start

```go
package main

import (
    "fmt"
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
}
```

## Configuration

```go
config := shrinkmap.Config{
    ShrinkInterval:        5 * time.Minute,  // How often to check for shrinking
    ShrinkRatio:          0.25,             // Ratio of deleted items that triggers shrinking
    InitialCapacity:      16,               // Initial map capacity
    AutoShrinkEnabled:    true,             // Enable automatic shrinking
    MinShrinkInterval:    30 * time.Second, // Minimum time between shrinks
    MaxMapSize:           1000000,          // Maximum map size before forcing shrink
    CapacityGrowthFactor: 1.2,             // Growth factor for new map allocation
}

sm := shrinkmap.New[string, int](config)
```

## Features in Detail

### Automatic Shrinking

The map automatically shrinks its internal storage when:
- The ratio of deleted items exceeds `ShrinkRatio`
- The time since the last shrink is greater than `MinShrinkInterval`
- `AutoShrinkEnabled` is true

```go
// Force immediate shrink
sm.ForceShrink()

// Try to shrink if conditions are met
sm.TryShrink()
```

### Performance Metrics

Monitor map performance with built-in metrics:

```go
metrics := sm.GetMetrics()
fmt.Printf("Total shrinks: %d\n", metrics.TotalShrinks)
fmt.Printf("Last shrink duration: %v\n", metrics.LastShrinkDuration)
fmt.Printf("Total items processed: %d\n", metrics.TotalItemsProcessed)
fmt.Printf("Peak size: %d\n", metrics.PeakSize)
```

### Thread Safety

All operations are thread-safe by default. The map can be safely used across multiple goroutines:

```go
// Safe for concurrent use
go func() {
    sm.Set("key1", 1)
}()

go func() {
    if value, exists := sm.Get("key1"); exists {
        // Handle value
    }
}()
```

### Range Operation

Iterate over all key-value pairs:

```go
sm.Range(func(key string, value int) bool {
    fmt.Printf("Key: %s, Value: %d\n", key, value)
    return true // Continue iteration
})
```

## Performance

Benchmark results compared to sync.Map (Go 1.21):

```
BenchmarkShrinkableMap/Set-8         5000000    234 ns/op    8 allocs/op
BenchmarkShrinkableMap/Get-8         20000000   60.1 ns/op   0 allocs/op
BenchmarkShrinkableMap/Delete-8      10000000   112 ns/op    0 allocs/op
```

## Best Practices

1. Choose appropriate shrink ratio:
```go
config.ShrinkRatio = 0.25 // Shrink when 25% of items are deleted
```

2. Adjust shrink interval based on usage patterns:
```go
config.ShrinkInterval = 1 * time.Minute // More frequent for high-churn maps
```

3. Set initial capacity for better performance:
```go
config.InitialCapacity = 1000 // When expecting ~1000 items
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Go's built-in map implementation
- Built with lessons learned from production systems
- Special thanks to all contributors

## Version History

- 0.0.1 (ing...)
    - Initial release
    - Generic type support
    - Automatic shrinking
    - Thread safety
    - Performance metrics

---
Made with ❤️ by [Jongyun Ha](https://github.com/jongyunha)