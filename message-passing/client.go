package messagepassing

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
)

type Client struct {
	connNumber int32

	name        string
	server      io.ReadWriteCloser
	serverRLock sync.Mutex
	serverWLock sync.Mutex

	newConnHandler NewConnectionhandler
	translator     MessageTranslator

	connections    map[string]*ClientConnections
	connectionLock sync.Mutex

	authed bool
}

type ClientConnections struct {
	isCLosed      uint32
	isEstablished uint32

	closed chan struct{}

	message         chan []byte
	currentMessaage []byte
	readLock        sync.Mutex

	OtherClient string
	connId      string
	client      *Client
}

func newClientConnection(otherClient, connID string, client *Client) *ClientConnections {
	return &ClientConnections{
		//message needs to remain unbuffered because of how our closedd channel works.
		message:     make(chan []byte),
		closed:      make(chan struct{}),
		OtherClient: otherClient,
		connId:      connID,
		client:      client,
	}
}

var (
	errWouldBlock = errors.New("Would block")
)

const (
	errSTringUnknownProtocol = "unknown protocol "
	errSTringMultuipleAuths  = "client has already authenticated"
	errStringNotYetAuthed    = "client not yet authenticated"
)

const (
	unreachableCode = "internal error --(supposedly) unreachable code was hit"
)

func (c *ClientConnections) ReadIntoBuffer(blocking bool) error {
	if c.currentMessaage != nil {
		return nil
	}
	var msg []byte
	if !blocking {
		select {
		case msg = <-c.message:
			break
		case <-c.closed:
			return io.EOF
		default:
			return errWouldBlock
		}
	} else {
		select {
		case msg = <-c.message:
			break
		case <-c.closed:
			return io.EOF
		}
	}
	c.currentMessaage = msg
	return nil
}

func (c *ClientConnections) IshanshakeComplete() bool {
	return atomic.LoadInt32(&c.isEstablished) != 0
}

func (c *Client) noteHandshakeComplete() {
	atomic.StoreUint32(&c.isEstablished, 1)
}

func (c *ClientConnections) Putmessage(msg *Message) bool {
	select {
	case <-c.closed:
		return false
	case c.message <- msg.Data:
		return true
	}
	panic(unreachableCode)
}

func (c *ClientConnections) Read(b []byte) (int, error) {
	c.readLock.Lock()
	defer c.readLock.Unlock()

	n := 0
	remslice := b
	for len(remslice) > 0 {
		block := n == 0
		err := c.ReadIntoBuffer(block)
		if err == errWouldBlock {
			return n, nil
		}

		x := copy(remslice, c.currentMessaage)
		if len(c.currentMessaage) == x {
			//release the message instead of letting it linger as a 0 size slice.
			c.currentMessaage = nil
		} else {
			c.currentMessaage = c.currentMessaage[x:]
		}
		remslice = remslice[x:]
		n += x
	}
	return n, nil
}

func (c *ClientConnections) Write(b []byte) (int, error) {
	if err := c.WriteMessage(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *ClientConnections) Close() error {
	if atomic.CompareAndSwapInt32(&c.isCLosed, 0, 1) {
		close(c.closed)
		//nil check beacuase testing can nil out the client
		if client := c.client; client != nil {
			client.notifyClosed(c)
		}
	}
	return nil
}

func (c *ClientConnections) CloseNotify() error {
	if atomic.CompareAndSwapInt32(&c.isCLosed, 0, 1) {
		close(c.closed)
	}
	return nil
}

func (c *ClientConnections) ReadMessage() ([]byte, error) {
	c.readLock.Lock()
	defer c.readLock.Unlock()

	err := c.ReadIntoBuffer(true)
	if err != nil {
		return nil, err
	}

	msg := c.currentMessaage
	c.currentMessaage = nil
	return msg, nil
}

func (c *ClientConnections) WriteMessage(b []byte) error {
	msg := Message{
		Meta:         MetaNone,
		OtherClient:  c.OtherClient,
		ConnectionID: c.connId,
		Data:         b,
	}

	err := c.client.sendMessage(&msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientConnections) OtherClient() string {
	return c.OtherClient
}

func NewClient(
	name string,
	server io.ReadWriteCloser,
	tm TranslatorMaker,
	ch NewConnectionhandler) *Client {
	return &Client{
		name:        name,
		server:      server,
		serverRLock: sync.Mutex{},
		serverWLock: sync.Mutex{},

		newConnHandler: ch,
		translator:     tm(server, server),

		connections:    make(map[string]*ClientConnections),
		connectionLock: sync.Mutex{},
	}
}

func (c *Client) sendMessage(m *Message) error {
	c.serverWLock.Lock()
	defer c.serverWLock.Unlock()

	return c.translator.WriteMessage(m)
}

func (c *Client) recvMessage() (*Message, error) {
	c.serverRLock.Lock()
	defer c.serverRLock.Unlock()

	return c.translator.ReadMessage()
}

// Authenticate allows the client to perform itsinitial handshake and
// authenticate itslef with a server. this should be called before
// Client.Run()
func (c *Client) Authenticate(password []byte) error {
	if c.authed {
		return errors.New(errSTringMultuipleAuths)
	}

	msg := Message{
		Meta:        MetaAuth,
		OtherClient: c.name,
		Data:        password,
	}

	err := c.sendMessage(&msg)
	if err != nil {
		return err
	}

	resp, err := c.recvMessage()
	if err != nil {
		return err
	}

	if resp.Meta != MetaAuth {
		return errors.New(string(resp.Data))
	}

	c.authed = true
	return nil
}

func (c *Client) NextConnIDNonAtomic() string {
	nextInt := c.connNumber
	c.connNumber++

	var microoptimization [64]byte
	buf := microoptimization[:0]
	buf = append(buf, c.name...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(nextInt), 16)
	return string(buf)
}

func (c *Client) MakeAndAddClientConn(otherClient string) (conn *ClientConnections, id string) {
	c.connectionLock()
	defer c.connectionLock.Unlock()

	for {
		id = c.NextConnIDNonAtomic()
		//clooocococoococoococ
		if _, ok := c.connections[id]; !ok {
			break
		}
	}

	conn = newClientConnection(otherClient, id, c)
	c.connections[id] = conn
	return
}

func (c *Client) MakeConnection(otherClient, proto string) (Connection, error) {
	conn, id := c.MakeAndAddClientConn(otherClient)

	msg := &Message{
		Meta:         MetaConnSyn,
		OtherClient:  otherClient,
		ConnectionID: id,
		Data:         []byte(proto),
	}

	err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	data, err := conn.ReadMessage()
	if err != nil {
		return nil, errors.New(string(data))
	}
	return conn, nil
}

func (c *Client) findAnyConnection(id string) (*ClientConnections, bool) {
	c.connectionLock.Lock()
	a, b := c.connections[id]
	c.connectionLock.Unlock()
	return a, b
}

func (c *Client) FindEstablishedConnection(id string) (*ClientConnections, bool) {
	a, ok := c.findAnyConnection(id)
	if !ok || !a.IshanshakeComplete() {
		return nil, false
	}
	return a, true
}

func (c *Client) removeConnection(id string) (*ClientConnections, bool) {
	c.connectionLock.Lock()
	conn, ok := c.connections[id]
	if ok {
		delete(c.connections, id)
	}
	c.connectionLock.Unlock()

	return conn, ok
}

// This was lifted out of handle MetaMessage because of its complexity
func (c *Client) MetaHandleSyn(msg *Message, proto string) (resp bool, er error) {
	var (
		conn                       *ClientConnections
		acceptCalled, callFineshed bool
	)

	accept := func() Connection {
		if callFineshed {
			panic("You need to call accept() before IncomingConnection() returns")
		}

		acceptCalled = true
		conn = newClientConnection(msg.OtherClient, msg.ConnectionID, c)

		// The whole reason that the complexity of having an accept() function
		// exists is so that we can guarantee that we send the ACK out before
		// the user gets their hands on the Connection object.

		c.connectionLock.Lock()
		c.connections[msg.ConnectionID] = conn
		c.connectionLock.Unlock()

		msg.Meta = MetaConnACk

		_ = c.sendMessage(msg)

		conn.noteHandshakeComplete()
		return conn
	}

	c.newConnHandler.IncomingConnection(proto, accept)
	callFineshed = true
	if !acceptCalled {
		msg.Meta = MetaUnknownProto
		return true, nil
	}
	return false, nil

}

// Close closes a Client's connection to the server, makes Client.Run() exit
// (eventually), and closes all Connection instances that this Client create
func (c *Client) Close() error {
	err := c.server.Close()

	c.connectionLock.Lock()
	connections := c.connections
	c.connections = make(map[string]*ClientConnections)
	c.connectionLock.Unlock()

	for _, conn := range connections {
		conn.CloseNotify()
	}
	return err
}

func (c *Client) handleMetaMessage(msg *Message) (resp bool, err error) {
	// In most cases, it's okay to bounce data back with our message, but it's
	// also undesirable to do so. So we nil out msg.Data and keep a snapshot of
	// what the data is, in case we explicitly want to send back the data that was
	// sent (or somehow use the data)
	msgData := msg.Data
	msg.Data = nil
	switch msg.Meta {
	case MetaNone:
		panic("Passed unmeta message to handleMetaMessage")
	case MetaNoSuchConnection, MetaConnClosed:
		conn, ok := c.removeConnection(msg.ConnectionID)
		if ok {
			conn.CloseNotify()
		}
		return false, nil
	case MetaUnknownProto:
		cid := msg.ConnectionID
		conn, ok := c.removeConnection(cid)

		if ok {
			// !!! It's assumed that conn.isHandshakeComplete() will *not* change
			// throughout the execution of this if statement. Otherwise, this msg
			// will leak to the client.
			// (It's also expected that we'll never get MetaUnknownProto if our
			// handshake is complete, but the extra protection doesn't hurt)
			if !conn.IshanshakeComplete() {
				msg.Data = []byte(errSTringUnknownProtocol)
				conn.Putmessage(msg)
			}
			conn.CloseNotify()
		}
		return false, nil
	case MetaConnSyn:
		proto := string(msgData)
		return c.MetaHandleSyn(msg, proto)
	case MetaConnACk:
		id := msg.ConnectionID
		conn, ok := c.findAnyConnection(id)

		if !ok {
			msg.Meta = MetaNoSuchConnection
			return true, nil
		}

		conn.noteHandshakeComplete()
		conn.Putmessage(msg)
		return false, nil
	case MetaClientCLosed:
		otherClient := msg.OtherClient
		connections := c.connections
		c.connectionLock.Lock()
		for k, conn := range c.connections {
			if conn.OtherClient() == otherClient {
				// If we want to change this to Close(), we need to move it out of this
				// loop. Otherwise, we'll hit a deadlock when Close() is trying to
				// notify the client of the closing (i.e. when it tries to acquire
				// connectionsLock)
				conn.CloseNotify()
				delete(connections, k)
			}
		}
		c.connectionLock.Unlock()
		return false, nil
	case MetaAuth, MetaAuthOk, MetaauthFailure:
		msg.Meta = MetaWAT
		return true, nil
	default:
		s := fmt.Sprintf("Unknown meta message type passed into handleMetaMessage: %d", msg.Meta)
		return false, errors.New(s)
	}

	// compat with Go 1.0
	panic(unreachableCode)
}

func (c *Client) notifyClosed(conn *ClientConnections) error {
	_, ok := c.removeConnection(conn.connId)
	if !ok {
		return nil
	}

	closedMsg := &Message{
		Meta:         MetaConnClosed,
		OtherClient:  conn.OtherClient,
		ConnectionID: conn.connId,
	}

	return c.sendMessage(closedMsg)
}

// Run is the main "event loop" for a Client. In it, the Client receives
// messages from the Server and dispatches them to Connections. You should
// call Authenticate() prior to calling Run().
func (c *Client) Run() error {
	if !c.authed {
		return errors.New(errStringNotYetAuthed)
	}

	defer c.server.Close()
	for {
		msg, err := c.recvMessage()
		if err != nil {
			return err
		}

		if msg.Meta == MetaNone {
			conn, ok := c.FindEstablishedConnection(msg.ConnectionID)
			if !ok || !conn.Putmessage(msg) {
				msg.Data = nil
				msg.Meta = MetaNoSuchConnection
				err = c.sendMessage(msg)
				if err != nil {
					return err
				}
			}
		} else {
			respond, err := c.handleMetaMessage(msg)
			if err != nil {
				return err
			}

			if respond {
				err = c.sendMessage(msg)
				if err != nil {
					return err
				}
			}
		}
	}
	// compat with Go 1.0
	panic(unreachableCode)
}
