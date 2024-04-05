# Lamport Clock 
Lamport Clock is a concept in distributed systems, introduced by computer scientist Leslie Lamport in 1978. It's a simple logical clock used to order events in a distributed system.

“Use logical timestamps as a version for a value to allow ordering of values across servers” 

## How it works:
- Each process in the system maintains its own Lamport Clock.
- Whenever an event occurs at a process, it increments its Lamport Clock by one.
- When a process sends a message to another process, it includes its current Lamport Clock value with the message.
- Upon receiving a message, the receiving process updates its Lamport Clock by taking the maximum of its current value and the timestamp received in the message, then increments it by one for the local event.
- This ensures that events are ordered according to the Lamport timestamps.
----
Lamport Clocks provide a partial ordering of events, meaning that if event A happened before event B according to Lamport timestamps, then event A definitely happened before event B. However, it doesn't guarantee that two events with different Lamport timestamps are actually causally related. For that, more sophisticated mechanisms like vector clocks or other causal ordering methods are needed.

## Example Usage:
Lamport Clocks can be applied in various scenarios within distributed systems to order events and maintain consistency despite the absence of a global clock.

**Message Passing**: Consider two processes, A and B, in a distributed system. Each process has its own Lamport Clock. When process A sends a message to process B, it includes its current ``Lamport Clock`` value with the message. Upon receiving the message, process B updates its ``Lamport Clock`` to be greater than the received timestamp and then processes the message. This ensures that the Lamport timestamps are used to order the events of sending and receiving messages between processes.

**Event Ordering**: Suppose we have three processes, P1, P2, and P3, in a distributed system. Each process generates local events and communicates with other processes through messages. If P1 generates an event at Lamport timestamp 5, and later P2 generates an event at Lamport timestamp 7, and finally, P3 generates an event at Lamport timestamp 6, the events will be ordered as follows: P1 event (timestamp 5) -> P3 event (timestamp 6) -> P2 event (timestamp 7). This is because the Lamport timestamps provide a partial ordering of events in the system.

**Concurrency Control**: In a distributed database system, multiple transactions may be executed concurrently across different nodes. Lamport Clocks can be used to order the execution of transactions to maintain consistency. Each transaction's start and end times can be recorded using Lamport timestamps, and this information can be used to ensure that transactions are executed in the correct order to maintain data consistency across the distributed system.

### Example  ``Go`` code 
- Just a quick ```main```
```
package main

import (
    "fmt"
    "sync"
)

type LamportClock struct {
    time int
    mu   sync.Mutex
}

func (lc *LamportClock) Tick() {
    lc.mu.Lock()
    lc.time++
    lc.mu.Unlock()
}

func (lc *LamportClock) Time() int {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    return lc.time
}

type Process struct {
    ID      int
    Clock   *LamportClock
    Channel chan int
}

func (p *Process) Event() {
    p.Clock.Tick()
    fmt.Printf("Process %d: Event occurred at time %d\n", p.ID, p.Clock.Time())
}

func (p *Process) SendEvent(receiver *Process) {
    p.Clock.Tick()
    receiver.Channel <- p.Clock.Time()
    fmt.Printf("Process %d: Sent message to Process %d at time %d\n", p.ID, receiver.ID, p.Clock.Time())
}

func (p *Process) ReceiveEvent(sender *Process) {
    msgTime := <-p.Channel
    p.Clock.Tick()
    p.Clock.time = max(p.Clock.time, msgTime) + 1
    fmt.Printf("Process %d: Received message from Process %d at time %d\n", p.ID, sender.ID, p.Clock.Time())
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func main() {
    // Create Lamport Clocks for processes
    clock1 := &LamportClock{}
    clock2 := &LamportClock{}

    // Create channels for message passing between processes
    channel1 := make(chan int)
    channel2 := make(chan int)

    // Create processes with their respective clocks and channels
    process1 := &Process{ID: 1, Clock: clock1, Channel: channel2}
    process2 := &Process{ID: 2, Clock: clock2, Channel: channel1}

    // Simulate events and message passing
    go process1.Event()
    go process2.SendEvent(process1)
    go process1.ReceiveEvent(process2)

    // Wait for goroutines to finish
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        wg.Wait()
    }()
    wg.Wait()
}
```