package messagepassing

import "io"

//Message is the type that client/server send over the wire
type Message struct{
	Meta MetaType 
	OtherClient string
	ConnectionID string
	Data []byte 
}

//metaType describes the intent of a message --some messgae are meant to simply
//be dilivered to a connection 
type MetaType uint8

//Constants o represent each Meta Type 
const (
	MetaNone = MetaType(iota)
	MetaWAT //Invalid/unkown meta type recieved 
	MetaNoSuchConnection 
	MetaUnknownProto
	MetaClientCLosed 
	MetaConnACk 
	MetaConnClosed 
	MetaAuth 
	MetaAuthOk 
	MetaauthFailure
)

type Connection interface{
	io.ReadWriteCloser

	//Reads the content of a single messages sent to us, blocking if necessary.
	ReadMessage() ([]byte, error)

	//Write a message, quarnteeing that the []byte given is the fill body 
	//of the message.
	WriteMessage([]byte) error

	//Gets the name of the client that our other Connection resides on.
	OtherClient(string, error)
}

//NewConnectionhandlers is an interface that client call when a request for 
//a new Connection comes in .
type NewConnectionhandler interface{
	// IncomingConnection. If you violate either of these, accept() panics.
	IncomingConnection(proto string, accept func() Connection)
}

//MessageTranslator is type that can read a message from a Reader and write 
//a message to a writer in some format.
type MessageTranslator interface{
	ReadMessage() (*Message, error)
	WriteMessage(*Message) error
}