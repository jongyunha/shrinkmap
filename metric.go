package shrinkmap

import (
	"sync"
	"time"
)

// ErrorRecord represents a single error or panic occurrence
type ErrorRecord struct {
	Timestamp time.Time
	Error     interface{} // panic 값이나 error 둘 다 저장 가능
	Stack     string      // 스택 트레이스 저장
}

// Metrics tracks performance and error metrics of the map
type Metrics struct {
	mu                  sync.RWMutex
	totalShrinks        int64
	lastShrinkDuration  time.Duration
	totalItemsProcessed int64
	peakSize            int32

	shrinkPanics  int64
	lastPanicTime time.Time
	lastError     *ErrorRecord
	errorHistory  []ErrorRecord
	totalErrors   int64
}

func (m *Metrics) TotalShrinks() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalShrinks
}

func (m *Metrics) LastShrinkDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastShrinkDuration
}

func (m *Metrics) TotalItemsProcessed() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalItemsProcessed
}

func (m *Metrics) PeakSize() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.peakSize
}

func (m *Metrics) RecordError(err error, stack string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := ErrorRecord{
		Timestamp: time.Now(),
		Error:     err,
		Stack:     stack,
	}

	m.lastError = &record
	m.totalErrors++

	if len(m.errorHistory) >= 10 {
		m.errorHistory = m.errorHistory[1:]
	}
	m.errorHistory = append(m.errorHistory, record)
}

func (m *Metrics) RecordPanic(r interface{}, stack string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := ErrorRecord{
		Timestamp: time.Now(),
		Error:     r,
		Stack:     stack,
	}

	m.lastError = &record
	m.shrinkPanics++
	m.lastPanicTime = time.Now()

	if len(m.errorHistory) >= 10 {
		m.errorHistory = m.errorHistory[1:]
	}
	m.errorHistory = append(m.errorHistory, record)
}

func (m *Metrics) TotalPanics() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shrinkPanics
}

func (m *Metrics) LastPanicTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastPanicTime
}

func (m *Metrics) LastError() *ErrorRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.lastError == nil {
		return nil
	}
	cp := *m.lastError
	return &cp
}

func (m *Metrics) ErrorHistory() []ErrorRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	history := make([]ErrorRecord, len(m.errorHistory))
	copy(history, m.errorHistory)
	return history
}

func (m *Metrics) TotalErrors() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalErrors
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalShrinks = 0
	m.lastShrinkDuration = 0
	m.totalItemsProcessed = 0
	m.peakSize = 0
	m.shrinkPanics = 0
	m.lastPanicTime = time.Time{}
	m.lastError = nil
	m.errorHistory = nil
	m.totalErrors = 0
}
