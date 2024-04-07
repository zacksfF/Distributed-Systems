package hybridclock

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

// Clock is a hybrid logical clock. Objects of this
// type model causality while maintaining a relation
// to physical time.
type Clock struct {
	physicalClock func() int64
	// Clock contains a mutex used to lock the below
	// fields while methods operate on them.
	state Timestamp
	// MaxOffset specifies how far ahead of the physical
	// clock (and cluster time) the wall time can be.
	// See SetMaxOffset.
	maxOffset time.Duration

	// monotonicityErrorsCount indicate how often this clock was
	// observed to jump backwards.
	monotonicityErrorsCount int32
	// lastPhysicalTime reports the last measured physical time. This
	// is used to detect clock jumps.
	lastPhysicalTime int64
}

// HybridClock represents a combined logical and physical clock
type HybridClock struct {
	logicalCounter int64
	physicalTime   time.Time
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

type Timestamp struct {
	counter int64
	time    int64
}

// ManualClock is a convenience type to facilitate
// creating a hybrid logical clock whose physical clock
// is manually controlled. ManualClock is thread safe.
type ManualClock struct {
	nanos int64
	time  time.Time
}

const (
	doubleQuote = "\""
	dash        = "-"
)

var (
	byteDash = []byte("-")
	//ErrWrongFormat is an error which only happens during UnmarshalJSOn
	//if and only if given data is not parsable to Timestamp
	ErrWrongFormat = errors.New("wrongformat")
)

// UnixNano returns the local machine's physical nanosecond
// unix epoch timestamp as a convenience to create a HLC via
// c := hlc.NewClock(hlc.UnixNano).
func UnixNano() int64 {
	return time.Now().UnixNano()
}

// NewManualClock returns a new instance, initialized with
// specified timestamp.
func NewManualClock(nanos int64) *ManualClock {
	return &ManualClock{nanos: nanos}
}

func (ts *Timestamp) String() string {
	return fmt.Sprintf("%s-%s", strconv.FormatInt(ts.time, 16), strconv.FormatInt(ts.counter, 16))
}

// UnixNano return the underlying manual clock's timestamp.
func (m *ManualClock) UnixNano() int64 {
	return atomic.LoadInt64(&m.nanos)
}

// Increment atomically increments the manual clock's timestamp.
func (m *ManualClock) Increment(incr int64) {
	atomic.AddInt64(&m.nanos, incr)
}

// Set atomically sets the manual clock's timestamp.
func (m *ManualClock) Set(nanos int64) {
	atomic.StoreInt64(&m.nanos, nanos)
}

// MarshalJSON  overrides and implements how timestamp needs to be encoded in JSON
func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString(doubleQuote)
	buffer.WriteString(ts.String())
	buffer.WriteString(doubleQuote)

	return buffer.Bytes(), nil
}

// UnmarshalJSON overrides and implement how timestamp should be parsed from JSOn
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	//need to remove quotes from data
	data = bytes.Trim(data, doubleQuote)

	segments := bytes.Split(data, byteDash)
	if len(segments) != 2 {
		return ErrWrongFormat
	}
	var err error
	ts.time, err = strconv.ParseInt(string(segments[0]), 16, 64)
	if err != nil {
		return err
	}
	ts.time, err = strconv.ParseInt(string(segments[1]), 16, 64)
	if err != nil {
		return err
	}
	return nil
}

// NewClock creates a new hybrid logical clock
func NewClock(physicalClock func() int64) *Clock {
	return &Clock{
		physicalClock: physicalClock,
	}
}

// // SetMaxOffset sets the maximal offset of the physical clock from the cluster.
// func (c *Clock) SetMaxOffset(delta time.Duration) {
// 	c.Lock()
// 	defer c.Unlock()
// 	c.maxOffset = delta
// }

// // MaxOffset returns the maximal offset allowed.
// // A value of 0 means offset checking is disabled.
// // See SetMaxOffset for details.
// func (c *Clock) MaxOffset() time.Duration {
// 	c.Lock()
// 	defer c.Unlock()
// 	return c.maxOffset
// }

// Less checks whether the given timestamp is bigger than current one
func (ts *Timestamp) Less(recv *Timestamp) bool {
	switch {
	case ts.time < recv.time:
		return true
	case ts.time == recv.time && ts.counter < recv.counter:
		return true
	default:
		return false
	}
}

// Now creates a new timestamp based on current clock
// This method should be called if sending a message or local change is occured
func (ts *Timestamp) Now() *Timestamp {
	t := ts.time
	ts.time = max2(t, pt())

	if ts.time == t {
		ts.counter++
	} else {
		ts.counter = 0
	}

	return &Timestamp{
		counter: ts.counter,
		time:    ts.time,
	}
}

func pt() int64 {
	return time.Now().UTC().UnixNano()
}

func max2(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}

func max3(a, b, c int64) int64 {
	if a < b {
		a = b
	}

	if a < c {
		a = c
	}

	return a
}

// Update the current clock, this should be called once a msg is recceived from
// other nodes
func (ts *Timestamp) Update(msg *Timestamp) {
	t := ts.time
	ts.time = max3(t, msg.time, pt())
	if ts.time == t && t == msg.time {
		ts.counter = max2(ts.counter, msg.counter) + 1
	} else if ts.time == t {
		ts.counter++
	} else if ts.time == msg.time {
		ts.counter = msg.counter + 1
	} else {
		ts.counter = 0
	}
}

// New creates a brand new Clock
// This function should be called once per node
func New() *Timestamp {
	return &Timestamp{}
}
