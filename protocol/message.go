package protocol

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

var (
	Err_BadFrame = errors.New("FrameIsIncorrect")
)

type Tag = uint16

type Message struct {
	Tag    Tag
	Length uint16
	Value  []byte
}

func read[T any](r io.Reader) (T, error) {
	var t T
	err := binary.Read(r, binary.LittleEndian, &t)

	return t, err
}

func Unmarshal(r io.Reader) (*Message, error) {
	tag, err := read[uint16](r)
	if err != nil {
		return nil, err
	}
	length, err := read[uint16](r)
	if err != nil {
		return nil, err
	}

	buff := make([]byte, length)
	n, err := bufio.NewReader(r).Read(buff)
	if err != nil {
		return nil, err
	}
	if n > len(buff) {
		return nil, Err_BadFrame
	}

	return &Message{
		Tag:    tag,
		Length: length,
		Value:  buff,
	}, nil
}

func (m *Message) Marshal(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.Tag); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.Length); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.Value); err != nil {
		return err
	}
	return nil
}
