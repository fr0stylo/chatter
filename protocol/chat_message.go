package protocol

import "fmt"

type ChatMessage struct {
	User    string
	Message string
}

func (c ChatMessage) String() string {
	return fmt.Sprintf("[%s] %s", c.User, c.Message)
}
