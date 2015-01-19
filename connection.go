// Package pusher implements client library for pusher.com socket
package pusher

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	pusherUrl = "ws://ws.pusherapp.com:80/app/%s?protocol=7"
)

type Connection struct {
	key      string
	conn     *websocket.Conn
	channels []*Channel

	mu     sync.Mutex
	closed bool
}

func New(key string) (*Connection, error) {
	ws, err := websocket.Dial(fmt.Sprintf(pusherUrl, key), "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	connection := &Connection{
		key:  key,
		conn: ws,
		channels: []*Channel{
			NewChannel(""),
		},
	}

	go connection.pong()
	go connection.poll()

	return connection, nil
}

func (c *Connection) pong() {
	tick := time.Tick(time.Minute)
	pong := NewPongMessage()
	for {
		<-tick
		if c.isClosed() {
			return
		}

		websocket.JSON.Send(c.conn, pong)
	}
}

func (c *Connection) poll() {
	for {
		if c.isClosed() {
			return
		}

		var msg Message
		err := websocket.JSON.Receive(c.conn, &msg)
		if err != nil {
			if c.isClosed() {
				return
			}

			log.Println("Pusher recieve error", err)
			<-time.After(time.Millisecond * 20)
			continue
		}

		c.processMessage(&msg)
	}
}

func (c *Connection) processMessage(msg *Message) {
	for _, channel := range c.channels {
		if channel.Name == msg.Channel {
			channel.processMessage(msg)
		}
	}
}

func (c *Connection) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}

func (c *Connection) Disconnect() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	return nil
}

func (c *Connection) Channel(name string) *Channel {
	for _, channel := range c.channels {
		if channel.Name == name {
			return channel
		}
	}

	channel := NewChannel(name)
	c.channels = append(c.channels, channel)
	websocket.JSON.Send(c.conn, NewSubscribeMessage(name))

	return channel
}
