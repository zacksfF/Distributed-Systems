package main

import (
	"fmt"
	"math/rand"
	"time"
)

// HybridClock represents a combined logical and physical clock
type HybridClock struct {
	logicalCounter int64
	physicalTime   time.Time
}

// NewHybridClock creates a new HybridClock instance
func NewHybridClock() *HybridClock {
	return &HybridClock{
		logicalCounter: 0,
		physicalTime:   time.Now(),
	}
}

// Increment increments the logical counter and adjusts for time drift
func (c *HybridClock) Increment() {
	c.logicalCounter++
	// Simulate time drift with a random delay (replace with actual drift detection)
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	newTime := time.Now()
	if newTime.Sub(c.physicalTime) > 100*time.Millisecond {
		c.logicalCounter++
		c.physicalTime = newTime
	}
}

// GetValue returns a tuple with logical counter and physical timestamp
func (c *HybridClock) GetValue() (int64, time.Time) {
	return c.logicalCounter, c.physicalTime
}

func main() {
	clock := NewHybridClock()

	// Simulate some events with increments
	for i := 0; i < 5; i++ {
		clock.Increment()
		fmt.Printf("Event %d happened at (logical: %d, physical: %v)\n", i+1, clock.GetValue())
	}
}
