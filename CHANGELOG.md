# Changelog

All notable changes to the ShrinkableMap project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.2] - 2024-11-02

### Added
- Error tracking and panic recovery system
    - Added ErrorRecord struct for detailed error information
    - Implemented error history with last 10 records
    - Added panic recovery statistics and tracking
- State inspection capabilities
    - Added Snapshot() method for safe state examination
    - Added KeyValue struct for type-safe iteration
- Resource management improvements
    - Added Stop() method for graceful shutdown
    - Implemented context-based goroutine management
- Enhanced metrics system
    - Added comprehensive error statistics
    - Added panic occurrence tracking
    - Added metrics reset capability
- Improved test coverage
    - Added error tracking tests
    - Added panic recovery tests
    - Added concurrent metrics access tests
    - Added snapshot functionality tests
    - Added resource cleanup tests

### Changed
- Improved thread safety with additional atomic operations
- Enhanced documentation with new examples and best practices
- Optimized memory usage in metrics collection

### Fixed
- Potential goroutine leak in auto-shrink feature
- Race conditions in metrics updates

## [0.0.1] - 2024-11-01

### Added
- Initial release
- Thread-safe implementation with atomic operations
- Generic type support
- Automatic shrinking with configurable policies
- Performance metrics tracking
- Comprehensive benchmark suite
- Basic configuration options
    - Shrink interval
    - Shrink ratio
    - Initial capacity
    - Auto-shrink settings
