// Package pusher implements client library for pusher.com socket
package pusher

import (
	"code.google.com/p/go.net/websocket"
)

type Client struct {
	key      string
	conn     *websocket.Conn
	channels []*Channel
}

func New(key string) (*Client, error) {
	url := "ws://ws.pusherapp.com:80/app/" + key

	ws, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	client := &Client{key, ws, nil}

	go func() {
		for {
			var msg Message
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				panic(err)
			}

			client.processMessage(&msg)
		}
	}()

	return client, nil
}

func (c *Client) processMessage(msg *Message) {
	switch msg.Event {
	case "pusher:connection_established":
		//println(msg.Data.(map[string]string)["socket_id"])
	case "connection_established":
		//println(msg.Data.(map[string]string)["socket_id"])
	case "pusher:error":
		//println(msg.Data.(map[string]interface{})["message"].(string))
	default:
		for _, channel := range c.channels {
			if channel.Name == msg.Channel {
				channel.processMessage(msg)
			}
		}
	}
}

func (c *Client) Channel(name string) *Channel {
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
