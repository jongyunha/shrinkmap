# ShrinkableMap

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/jongyunha/shrinkmap.svg)](https://pkg.go.dev/github.com/jongyunha/shrinkmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/jongyunha/shrinkmap)](https://goreportcard.com/report/github.com/jongyunha/shrinkmap)
[![Coverage Status](https://coveralls.io/repos/github/jongyunha/shrinkmap/badge.svg?branch=main)](https://coveralls.io/github/jongyunha/shrinkmap?branch=main)

ShrinkableMap is a high-performance, generic, thread-safe map implementation for Go that automatically manages memory by shrinking its internal storage when items are deleted. It addresses the common issue where Go's built-in maps don't release memory after deleting elements.

## 🚀 Features

- **Type Safety**
    - Generic type support for compile-time type checking
    - Type-safe operations for all map interactions

- **Performance**
    - Optimized concurrent access with minimal locking
    - Efficient atomic operations for high throughput
    - Batch operations for improved performance

- **Memory Management**
    - Automatic memory shrinking with configurable policies
    - Advanced concurrent shrinking behavior
    - Memory-efficient iterators

- **Reliability**
    - Thread-safe implementation
    - Panic recovery and error tracking
    - Comprehensive metrics collection

- **Developer Experience**
    - Safe iteration with snapshot support
    - Batch operations for bulk processing
    - Clear error reporting and metrics
    - Zero external dependencies
    - Production-ready with extensive tests

## 📦 Installation

```bash
go get github.com/jongyunha/shrinkmap
```

## 🔧 Quick Start

```go
package main

import (
    "fmt"
    "github.com/jongyunha/shrinkmap"
)

func main() {
    // Create a new map with default configuration
    sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())
    defer sm.Stop()

    // Basic operations
    sm.Set("one", 1)
    sm.Set("two", 2)

    if value, exists := sm.Get("one"); exists {
        fmt.Printf("Value: %d\n", value)
    }

    // Delete value
    sm.Delete("one")
}
```

## 💡 Advanced Usage

### Batch Operations

Efficiently process multiple operations atomically:

```go
batch := shrinkmap.BatchOperations[string, int]{
    Operations: []shrinkmap.BatchOperation[string, int]{
        {Type: shrinkmap.BatchSet, Key: "one", Value: 1},
        {Type: shrinkmap.BatchSet, Key: "two", Value: 2},
        {Type: shrinkmap.BatchDelete, Key: "three"},
    },
}

// Apply all operations atomically
sm.ApplyBatch(batch)
```

### Safe Iteration

Iterate over map contents safely using the iterator:

```go
// Create an iterator
iter := sm.NewIterator()

// Iterate over all items
for iter.Next() {
    key, value := iter.Get()
    fmt.Printf("Key: %v, Value: %v\n", key, value)
}

// Or use snapshot for bulk processing
snapshot := sm.Snapshot()
for _, kv := range snapshot {
    fmt.Printf("Key: %v, Value: %v\n", kv.Key, kv.Value)
}
```

### Performance Monitoring

Track performance metrics:

```go
metrics := sm.GetMetrics()
fmt.Printf("Total operations: %d\n", metrics.TotalItemsProcessed())
fmt.Printf("Peak size: %d\n", metrics.PeakSize())
fmt.Printf("Total shrinks: %d\n", metrics.TotalShrinks())
```

## 🔍 Configuration Options

```go
config := shrinkmap.Config{
    InitialCapacity:      1000,
    AutoShrinkEnabled:    true,
    ShrinkInterval:       time.Second,
    MinShrinkInterval:    time.Second,
    ShrinkRatio:         0.5,
    CapacityGrowthFactor: 1.5,
    MaxMapSize:           1000000,
}
```

## 🛡️ Thread Safety Guarantees

- All map operations are atomic and thread-safe
- Safe concurrent access from multiple goroutines
- Thread-safe batch operations
- Safe iteration with consistent snapshots
- Coordinated shrinking operations
- Thread-safe metrics collection

## 📊 Performance

Benchmark results on typical operations (Intel i7-9700K, 32GB RAM):

```
BenchmarkBasicOperations/Sequential_Set-8         5000000    234 ns/op
BenchmarkBasicOperations/Sequential_Get-8         10000000   112 ns/op
BenchmarkBatchOperations/BatchSize_100-8         100000     15234 ns/op
BenchmarkConcurrency/Parallel_8-8                1000000    1123 ns/op
```

## 📝 Best Practices

1. **Resource Management**
```go
sm := shrinkmap.New[string, int](config)
defer sm.Stop() // Always ensure proper cleanup
```

2. **Batch Processing**
```go
// Use batch operations for multiple updates
batch := prepareBatchOperations()
sm.ApplyBatch(batch)
```

3. **Safe Iteration**
```go
// Use iterator for safe enumeration
iter := sm.NewIterator()
for iter.Next() {
    // Process items safely
}
```

## 🗓️ Version History

### v0.0.4 (Current)
- Added batch operations support for atomic updates
- Implemented safe iterator pattern
- Enhanced performance for bulk operations
- Added comprehensive benchmarking suite
- Improved documentation and examples

### history
#### v0.0.3
- enhanced performance and memory management
#### v0.0.2
- Added error tracking and panic recovery
- Added state snapshot functionality
- Added graceful shutdown
- Enhanced metrics collection
- Improved resource cleanup
#### v0.0.1
- Initial release with core functionality
- Thread-safe implementation
- Automatic shrinking support
- Generic type support

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
Made with ❤️ by [Jongyun Ha](https://github.com/jongyunha)
