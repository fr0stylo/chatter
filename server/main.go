package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fr0stylo/chateh/protocol"
)

type Connections = map[string]net.Conn

var connections = Connections{}

type writer chan []byte

func (w writer) Write(p []byte) (int, error) {
	w <- append(([]byte)(nil), p...)
	return len(p), nil
}

func writePump(w writer) {
	bw := bufio.NewWriter(os.Stderr)
	for p := range w {
		bw.Write(p)

		// Slurp up buffered messages in flush. This ensures
		// timely output.
		n := len(w)
		for i := 0; i < n; i++ {
			bw.Write(<-w)
		}
		bw.Flush()
	}
}

func main() {
	w := make(writer, 128) // adjust capacity to meet your needs
	go writePump(w)
	log.SetOutput(w)
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Ltime)
	l, err := net.Listen("tcp", "0.0.0.0:9494")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	msgC := make(chan *protocol.Message)

	go messageBroadcaster(ctx, msgC)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				c, err := l.Accept()
				if errors.Is(err, net.ErrClosed) {
					return
				} else if err != nil {
					log.Print(err)
					continue
				}
				go handleConnection(ctx, c, msgC)
			}
		}
	}()

	<-exit
	cancel()
	log.Print("Closing")
}

func handleConnection(ctx context.Context, rw net.Conn, messageBroadcaster chan *protocol.Message) {
	msgC := initiateReader(ctx, rw)
	c := time.NewTimer(15 * time.Second)
	id := RandStringBytes(6)
	for {
		select {
		case <-ctx.Done():
			rw.SetDeadline(time.Now())
			return
		case msg := <-msgC:
			switch msg.Tag {
			case protocol.Tag_Connected:
				log.Print("Client connected")
				connections[id] = rw

				go pinger(c, rw)
			case protocol.Tag_Disconnected:
				log.Print("Client disconnected")

				delete(connections, id)
				c.Stop()
				return
			case protocol.Tag_Pong:
				log.Print("PONG")
				c.Reset(15 * time.Second)
			case protocol.Tag_Message:
				var cm protocol.ChatMessage
				gob.NewDecoder(bytes.NewBuffer(msg.Value)).Decode(&cm)
				cm.User = id
				messageBroadcaster <- protocol.TagMessage(cm)
				// go broadcastMessage(protocol.TagMessage(cm))
			}

			log.Print(msg)
		}
	}
}

func pinger(t *time.Timer, rw io.ReadWriter) {
	for range t.C {
		protocol.TagPing.Marshal(rw)
	}
}

func initiateReader(ctx context.Context, r io.Reader) chan *protocol.Message {
	messageC := make(chan *protocol.Message, 20)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := protocol.Unmarshal(r)
				if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF) {
					return
				} else if err != nil {
					log.Print(err)
					return
				}

				messageC <- msg
			}
		}
	}()

	return messageC
}

func messageBroadcaster(ctx context.Context, messageC chan *protocol.Message) {
	for {
		select {
		case message := <-messageC:
			for _, v := range connections {
				message.Marshal(v)
			}
		case <-ctx.Done():
			return
		}
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
