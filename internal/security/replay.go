package security

import (
	"sync"
	"time"
)

// ReplayProtection prevents message replay attacks
type ReplayProtection struct {
	window    map[int64]time.Time
	windowSize int
	mu        sync.Mutex
}

// NewReplayProtection creates a new replay protection
func NewReplayProtection(windowSize int) *ReplayProtection {
	if windowSize <= 0 {
		windowSize = 100
	}
	return &ReplayProtection{
		window:    make(map[int64]time.Time),
		windowSize: windowSize,
	}
}

// Check checks if a message ID is valid (not replayed)
func (r *ReplayProtection) Check(msgID int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if in window
	if _, ok := r.window[msgID]; ok {
		return false // Already seen
	}

	// Add to window
	r.window[msgID] = time.Now()

	// Cleanup old entries if window is full
	if len(r.window) > r.windowSize {
		r.cleanup()
	}

	return true
}

// cleanup removes old entries
func (r *ReplayProtection) cleanup() {
	cutoff := time.Now().Add(-5 * time.Minute)
	for id, t := range r.window {
		if t.Before(cutoff) {
			delete(r.window, id)
		}
	}
}

// Clear clears the window
func (r *ReplayProtection) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.window = make(map[int64]time.Time)
}

// MessageCounter tracks message IDs with timestamps
type MessageCounter struct {
	mu       sync.Mutex
	lastID   int64
	lastTime time.Time
}

// NewMessageCounter creates a new counter
func NewMessageCounter() *MessageCounter {
	return &MessageCounter{}
}

// Next returns the next message ID (must be > last)
func (c *MessageCounter) Next() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if now.After(c.lastTime.Add(time.Millisecond)) {
		c.lastID = now.UnixMilli() << 32
	}
	c.lastID++

	return c.lastID
}

// RecentMessages tracks recent message IDs
type RecentMessages struct {
	messages map[int64]time.Time
	mu       sync.RWMutex
}

// NewRecentMessages creates a recent messages tracker
func NewRecentMessages() *RecentMessages {
	return &RecentMessages{
		messages: make(map[int64]time.Time),
	}
}

// Add adds a message ID
func (r *RecentMessages) Add(msgID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages[msgID] = time.Now()
}

// Exists checks if message ID exists
func (r *RecentMessages) Exists(msgID int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.messages[msgID]
	return ok
}

// Cleanup removes old messages
func (r *RecentMessages) Cleanup(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, t := range r.messages {
		if t.Before(cutoff) {
			delete(r.messages, id)
		}
	}
}
