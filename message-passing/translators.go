package messagepassing

import (
	"encoding/gob"
	"encoding/json"
	"io"
)

// Translates messages using the built-in encoding/gob module
type gobTranslator struct {
	dec *gob.Decoder
	enc *gob.Encoder
}

type jsonTranslators struct {
	dec *json.Decoder
	enc *json.Encoder
}

// NewGobTranslator creates a new MessageTranslator that reads/wrote messages
// Using the Gob message format.
func NewGobTranslator(r io.Reader, w io.Writer) MessageTranslator {
	return &gobTranslator{gob.NewDecoder(r), gob.NewEncoder(w)}
}

func (t *gobTranslator) ReadMessage() (*Message, error) {
	msg := &Message{}
	err := t.dec.Decode(msg)
	return msg, err
}

func (t *gobTranslator) WriteMessage(m *Message) error {
	return t.enc.Encode(m)
}

// //NewJSOnTranslators creates a new MessageTranslator that reads/write messages
// using the JSOn message format.
func NewJSOnTranslators(r io.Reader, w io.Writer) MessageTranslator {
	return &jsonTranslators{json.NewDecoder(r), json.NewEncoder(w)}
}

func (t *jsonTranslators) ReadMessage() (*Message, error) {
	msg := &Message{}
	err := t.dec.Decode(msg)
	return msg, err
}

func (t *jsonTranslators) WriteMessage(m *Message) error {
	return t.enc.Encode(m)
}
