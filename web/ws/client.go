package ws

import (
	"context"
	"time"

	"nhooyr.io/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second
	maxMsgSize = 4096
	sendBufLen = 256
)

// Client represents a WebSocket client connection.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// NewClient creates a new WebSocket client.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, sendBufLen),
	}
}

// WritePump sends messages from the send channel to the WebSocket connection.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Write(writeCtx, websocket.MessageText, message)
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// ReadPump reads messages from the WebSocket (mainly to detect disconnects).
func (c *Client) ReadPump(ctx context.Context) {
	c.conn.SetReadLimit(maxMsgSize)
	for {
		_, _, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
	}
}

// Close closes the WebSocket connection and send channel.
func (c *Client) Close() {
	c.conn.Close(websocket.StatusNormalClosure, "")
	close(c.send)
}
