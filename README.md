# ShrinkableMap

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

ShrinkableMap is a high-performance, generic, thread-safe map implementation for Go that automatically manages memory by shrinking its internal storage when items are deleted. It provides a solution to the common issue where Go's built-in maps don't release memory after deleting elements.

## Features

- 🚀 Generic type support for type-safe operations
- 🔒 Thread-safe implementation with atomic operations
- 📉 Automatic memory shrinking with configurable policies
- ⚙️ Advanced concurrent shrinking behavior
- 📊 Thread-safe performance and error metrics
- 🛡️ Panic recovery and error tracking
- 🔍 Safe state inspection with snapshots
- 🧹 Graceful resource cleanup
- 💪 Production-ready with comprehensive tests
- 🎯 Zero external dependencies

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

    // Ensure cleanup when done
    defer sm.Stop()

    // Set values
    sm.Set("one", 1)
    sm.Set("two", 2)

    // Get value
    if value, exists := sm.Get("one"); exists {
        fmt.Printf("Value: %d\n", value)
    }

    // Delete value
    sm.Delete("one")

    // Get current state snapshot
    snapshot := sm.Snapshot()
    for _, kv := range snapshot {
        fmt.Printf("Key: %v, Value: %v\n", kv.Key, kv.Value)
    }

    // Get metrics including error statistics
    metrics := sm.GetMetrics()
    fmt.Printf("Total operations: %d\n", metrics.TotalItemsProcessed())
    fmt.Printf("Total errors: %d\n", metrics.TotalErrors())
    fmt.Printf("Total panics: %d\n", metrics.TotalPanics())
}
```

## Advanced Features

### Error Tracking and Recovery

Monitor and track errors with detailed information:

```go
metrics := sm.GetMetrics()

// Get error statistics
totalErrors := metrics.TotalErrors()
totalPanics := metrics.TotalPanics()
lastPanicTime := metrics.LastPanicTime()

// Get last error details
if lastError := metrics.LastError(); lastError != nil {
    fmt.Printf("Last error: %v\n", lastError.Error)
    fmt.Printf("Stack trace: %v\n", lastError.Stack)
    fmt.Printf("Time: %v\n", lastError.Timestamp)
}

// Get error history (last 10 errors)
errorHistory := metrics.ErrorHistory()
for _, err := range errorHistory {
    fmt.Printf("Error: %v, Time: %v\n", err.Error, err.Timestamp)
}
```

### State Inspection

Safely inspect map state without locking:

```go
// Get current state snapshot
snapshot := sm.Snapshot()
for _, kv := range snapshot {
    fmt.Printf("Key: %v, Value: %v\n", kv.Key, kv.Value)
}
```

### Resource Management

Proper cleanup with graceful shutdown:

```go
// Create map
sm := shrinkmap.New[string, int](shrinkmap.DefaultConfig())

// Ensure cleanup
defer sm.Stop()

// Or stop explicitly when needed
sm.Stop()
```

## Thread Safety Guarantees

- All map operations are atomic and thread-safe
- Metrics collection is non-blocking and thread-safe
- Shrinking operations are coordinated to prevent conflicts
- Safe concurrent access from multiple goroutines
- Panic recovery in auto-shrink goroutine
- Thread-safe error tracking and metrics collection
- Safe state inspection with snapshots

## Best Practices

1. Always ensure proper cleanup:
```go
sm := shrinkmap.New[string, int](config)
defer sm.Stop() // Ensure auto-shrink goroutine is cleaned up
```

2. Monitor errors and panics:
```go
metrics := sm.GetMetrics()
if metrics.TotalErrors() > 0 {
    // Investigate error history
    for _, err := range metrics.ErrorHistory() {
        log.Printf("Error: %v, Time: %v\n", err.Error, err.Timestamp)
    }
}
```

3. Use snapshots for safe iteration:
```go
snapshot := sm.Snapshot()
for _, kv := range snapshot {
    // Process items without holding locks
    process(kv.Key, kv.Value)
}
```

## Version History

- 0.0.2 (ing...)
    - Added error tracking and panic recovery
    - Added state snapshot functionality
    - Added graceful shutdown with Stop()
    - Enhanced metrics with error statistics
    - Improved resource cleanup
    - Added comprehensive error tracking tests

- 0.0.1
    - Initial release
    - Thread-safe implementation with atomic operations
    - Generic type support
    - Automatic shrinking with configurable policies
    - Comprehensive benchmark suite
    - Race condition free guarantee

---
Made with ❤️ by [Jongyun Ha](https://github.com/jongyunha)
