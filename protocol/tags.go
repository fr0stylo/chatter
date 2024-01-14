package protocol

import (
	"bytes"
	"encoding/gob"
)

const (
	Tag_Connected Tag = iota
	Tag_Disconnected
	Tag_Ping
	Tag_Pong
	Tag_Message
)

var TagConnected = &Message{Tag: Tag_Connected, Length: 0}
var TagDisconnected = &Message{Tag: Tag_Disconnected, Length: 0}
var TagPing = &Message{Tag: Tag_Ping, Length: 0}
var TagPong = &Message{Tag: Tag_Pong, Length: 0}

func TagMessage(m ChatMessage) *Message {
	buf := bytes.NewBuffer(make([]byte, 0))
	gob.NewEncoder(buf).Encode(m)

	return &Message{
		Tag:    Tag_Message,
		Length: uint16(buf.Len()),
		Value:  buf.Bytes(),
	}
}
