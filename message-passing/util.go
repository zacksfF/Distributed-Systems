package messagepassing

// MappedConnectionHandler is a NewConnectionHandler that keeps an internal
// mapping ofprotocol -> function
type MappedConnectionHandler struct {
	chMap map[string]func(Connection)
}

// AddMapping is a maps protocol name to a function to be run in a goroutine.
func (m *MappedConnectionHandler) AddMapping(str string, fn func(Connection)) {
	m.chMap[str] = fn
}

// IncomingConnection allows us to implement the NewConnectionHandler interface
func (m *MappedConnectionHandler) IncomingConnection(proto string, accept func()Connection){
	c, ok := m.chMap[proto]
	if ok {
		go c(accept())
	}
}

//NewMappedConnectionHandler Creates and initializes a new instance of 
//MappedConnectionHandler
func NewConnectionHandler() *MappedConnectionHandler{
	return &MappedConnectionHandler{make(map[string]func(Connection))}
}