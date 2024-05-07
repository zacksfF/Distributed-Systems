package clientserver

import (
	"sync"
)

const (
	START_STREAM_ID_OF_CLIENT = 1
	START_STREAM_ID_OF_SERVER = 2
	MAX_STREAM_ID             = 1 << 31 // 2,147,483,648

	NUM_BYTES_HEADER      = 8
	NUM_BYTES_MAX_PAYLOAD = 1 << 14 // 16,384 bytes = 16kb

	TYPE_DATA = 0

	FLAG_DATA_NONE       = 0
	FLAG_DATA_END_STREAM = 1
)

type Client struct {
	sync.Mutex

	Network    string
	Address    string
	connection *Conn
}

func (c *Client) GetConn(force bool) (*Conn, error) {
	if c.connection == nil || force {
		conn, err := Dial(c.connection, c.Address)
		if err != nil {
			return nil, err
		}
		if c.connection != nil {
			c.connection.Close()
		}
		go connection.Listen()
		c.connection = &connection
	}
	return c.connection, nil
}
