package clientserver

import (
	"io"
	"net"
	"sync"
)

type Conn struct {
	net.Conn
	sync.Mutex
	streams sync.Map
	counter *Counter
	ch      chan Stream
}

type ConnInfo struct {
	c   net.Conn
	err error
}

type Counter struct {
	sync.Mutex
	current uint32
}

type Stream struct {
	conn   net.Conn
	id     uint32
	writer *io.PipeWriter
	reader *io.PipeReader
	in     chan []byte
}

func NewStream(id uint32, in chan []byte, c *Conn) Stream {
	pr, pw := io.Pipe()
	return Stream{
		id:     id,
		in:     in,
		conn:   c,
		reader: pr,
		writer: pw,
	}
}
func (c *Conn) Get(g uint)

func (c *Conn) Stream() (Stream, error) {
	stream := make(chan []byte, 10)
	id, err := c.counter.Get()
	if err != nil {
		return NewStream(id, stream, c), err
	}
	c.streams.Store(id, stream)
	return NewStream(id, stream, c), nil
}

func NewCounter(init uint32) *Counter {
	return &Counter{
		current: init,
	}
}

// Dial
func Dial(network, address string) (Conn, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return Conn{}, err
	}
	return Conn{
		Conn:    conn,
		streams: sync.Map{},
		ch:      make(chan Stream, 1),
		counter: NewCounter(START_STREAM_ID_OF_CLIENT),
	}, nil
}
