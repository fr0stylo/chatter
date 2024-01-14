package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/fr0stylo/chateh/protocol"
)

// type writer chan []byte

// func (w writer) Write(p []byte) (int, error) {
// 	w <- append(([]byte)(nil), p...)
// 	return len(p), nil
// }

// func writePump(w writer) {
// 	bw := bufio.NewWriter(os.Stderr)
// 	for p := range w {
// 		bw.Write(p)

// 		// Slurp up buffered messages in flush. This ensures
// 		// timely output.
// 		n := len(w)
// 		for i := 0; i < n; i++ {
// 			bw.Write(<-w)
// 		}
// 		bw.Flush()
// 	}
// }

// func main() {
// 	// w := make(writer, 128) // adjust capacity to meet your needs
// 	// go writePump(w)
// 	// log.SetOutput(w)
// 	// log.SetFlags(log.Lshortfile | log.LstdFlags | log.Ltime)

// 	con, err := net.Dial("tcp", "0.0.0.0:9494")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer con.Close()

// 	ctx, cancel := context.WithCancel(context.Background())
// 	protocol.TagConnected.Marshal(con)
// 	go handleInbound(ctx, con)

// 	exit := make(chan os.Signal, 1)
// 	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
// 	<-exit
// 	cancel()

// 	protocol.TagDisconnected.Marshal(con)
// 	log.Print("Closing")
// }

type Client struct {
	conn         net.Conn
	messageRecvC chan *protocol.ChatMessage
	messageSendC chan string
}

func (c *Client) Close() {
	protocol.TagDisconnected.Marshal(c.conn)
	c.conn.Close()
}

func (c *Client) SendChannel() chan string {
	return c.messageSendC
}

func (c *Client) ReceiveChannel() chan *protocol.ChatMessage {
	return c.messageRecvC
}

func Connect(ctx context.Context, ip string) *Client {
	con, err := net.Dial("tcp", ip)
	if err != nil {
		panic(err)
	}

	protocol.TagConnected.Marshal(con)
	msgRecvC := make(chan *protocol.ChatMessage)
	msgSendC := make(chan string)
	go handleInbound(ctx, con, msgRecvC)
	go handleOutbound(ctx, con, msgSendC)

	return &Client{
		conn:         con,
		messageRecvC: msgRecvC,
		messageSendC: msgSendC,
	}
}

func handleOutbound(ctx context.Context, rw net.Conn, messageSendC chan string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-messageSendC:
			if err := protocol.TagMessage(protocol.ChatMessage{
				Message: msg,
			}).Marshal(rw); err != nil {
				return err
			}
		}
	}
}

func handleInbound(ctx context.Context, rw net.Conn, messageRecvC chan *protocol.ChatMessage) error {
	msgC := initiateReader(ctx, rw)
	for {
		select {
		case <-ctx.Done():
			rw.SetReadDeadline(time.Now())

			return nil
		case msg := <-msgC:
			switch msg.Tag {
			case protocol.Tag_Ping:
				protocol.TagPong.Marshal(rw)
			case protocol.Tag_Message:
				var cm protocol.ChatMessage
				gob.NewDecoder(bytes.NewBuffer(msg.Value)).Decode(&cm)
				messageRecvC <- &cm
			}
		}
	}
}

func initiateReader(ctx context.Context, r io.Reader) chan *protocol.Message {
	messageC := make(chan *protocol.Message, 20)

	go func() {
		for {
			msg, err := protocol.Unmarshal(r)
			if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				log.Print(err)
				return
			}

			messageC <- msg
		}
	}()

	return messageC
}
