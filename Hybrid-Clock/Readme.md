# Hybrid Clock
A ``Hybrid Clock`` is a synchronization mechanism used in distributed systems to combine the benefits of both logical clocks (such as Lamport Clocks) and physical clocks (such as system clocks). It aims to provide both causality tracking and real-time ordering of events in a distributed environment.

``Hybrid Clock`` “Use a combination of system timestamp and logical timestamp to have versions as date-time, which can be ordered”
``Hybrid Logical`` Clock provides a way to have a version which is monotonically increasing just like a simple integer but also has a relation with the actual date time. Hybrid clocks are used in practice by databases like MongoDB or cockroach DB 

1. **Logical Clock Component**: Similar to Lamport Clocks, each process maintains a logical clock that increments with local events and message exchanges. This logical clock helps establish the causal relationship between events.
2. **Physical Clock Component**: Each process also maintains its own physical clock, which represents real-world time according to the local system's clock. The physical clock provides an approximation of real-time ordering.
3. **Synchronization**: Periodically, the physical clocks of different processes may need to be synchronized to prevent significant clock drift. This can be achieved using clock synchronization protocols like NTP (Network Time Protocol) or PTP (Precision Time Protocol).
4. **Combination Algorithm**: The Hybrid Clock combines both logical and physical components to generate timestamps for events. When a process needs to timestamp an event, it generates a hybrid timestamp by combining its logical clock value with the current value of its physical clock. This combination algorithm ensures that the hybrid timestamps reflect both the causality of events and the real-time ordering.

**How it Works:**

A Hybrid Clock is typically a tuple containing two values:

- **Physical Timestamp:** This is the reading from the machine's physical clock.
- **Logical Counter:** This is a monotonically increasing integer that keeps track of the event order within a node.

Whenever an event occurs, the Hybrid Clock is incremented:

1. The logical counter is incremented by 1.
2. The physical timestamp is compared with the current time on the machine's clock.
3. If the physical time has advanced significantly (due to drift), the logical counter is adjusted to ensure it stays ahead of physical time.

By combining logical and physical clocks, Hybrid Clocks aim to provide a more accurate and meaningful representation of event ordering in distributed systems, balancing both causality tracking and real-time considerations.

Would you like to see an example program demonstrating the use of a Hybrid Clock in Go, similar to the Lamport Clock example?

**Use Cases:**

Hybrid Clocks are used in various distributed system applications, including:

- **Distributed Databases:** They help maintain version history and ensure data consistency during concurrent operations.
- **Event Ordering:** They are crucial for systems where the order of events is critical, like message queues or distributed task processing.

**Benefits of Hybrid Clocks:**

- **Consistent Ordering:** They provide a way to order events across the system, even if physical clocks are not perfectly synchronized.
- **Real-world Context:** They allow referencing events with real-world timestamps for better understanding.
- **Versioning:** They're useful for versioning data in distributed systems, providing context for historical changes.

## A Hybrid Logical Clock is implemented as follows: 
Here's an example of a very basic Hybrid Clock implementation in Go:

```go
package main

import (
  "fmt"
  "time"
)

// HybridClock represents a combined logical and physical clock
type HybridClock struct {
  logicalCounter int64
  physicalTime  time.Time
}

// NewHybridClock creates a new HybridClock instance
func NewHybridClock() *HybridClock {
  return &HybridClock{
    logicalCounter: 0,
    physicalTime:  time.Now(),
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
```
