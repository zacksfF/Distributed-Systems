package clientserver

import (
	"net"
	"sync"
)


//listener hold the tcpSorter net.Listener interface implementation 
type Listener struct {
	prefixes [][]byte
	p        *Portal
	listn    chan ConnInfo
}

// Portal hold the port-listenning TCP sorter
type Portal struct {
	list     net.Listener
	failOver chan ConnInfo
	mu sync.Mutex
	listeners map[string]chan ConnInfo
}

