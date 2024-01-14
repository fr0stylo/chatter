package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestIsProtocolWorking(t *testing.T) {
	msg := &Message{Tag: 1, Length: 41, Value: []byte("asdfa sadf asdfasdfsadf asdf asdf asd fa!")}

	buf := bytes.NewBufferString("")

	if err := msg.Marshal(buf); err != nil {
		t.Errorf("Failed to marshal, %s", err)
	}

	recv, err := Unmarshal(buf)
	if err != nil {
		t.Errorf("Failed to unmarshal, %s", err)
	}

	if !reflect.DeepEqual(msg, recv) {
		t.Errorf("Entities are different, %v != %v", recv, msg)
	}
}

func BenchmarkMarshal(b *testing.B) {
	m := &Message{
		Tag:    Tag_Connected,
		Length: 60,
		Value:  []byte("01234567890123456789012  3456 789012345678901234567890123456789"),
	}

	buf := bytes.NewBuffer([]byte{})

	for i := 0; i < b.N; i++ {
		m.Marshal(buf)
		buf.Reset()
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	m := &Message{
		Tag:    Tag_Connected,
		Length: 60,
		Value:  []byte("012345678901234567890123456789012345678901234567890123456789"),
	}

	buf := bytes.NewBuffer([]byte{})
	m.Marshal(buf)

	for i := 0; i < b.N; i++ {
		Unmarshal(buf)
	}
}
