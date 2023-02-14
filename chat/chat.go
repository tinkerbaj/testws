package chat

import (
	"context"
	"fmt"

	// "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	// Time allowed to write a message to the peer.
	// writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// maxMessageSize = 512
)

// Client struct for websocket connection and message sending
type Client struct {
	ID   string
	Conn *websocket.Conn
	send chan Message
	hub  *Hub
}

// NewClient creates a new client
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{Conn: conn, send: make(chan Message, 256), hub: hub}
}

// Client goroutine to read messages from client
func (c *Client) Read() {

	defer func() {
		c.hub.unregister <- c
		c.Conn.Close(103, "Disconnect")
	}()

	for {
		var msg Message
		err := wsjson.Read(context.Background(), c.Conn, &msg)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		c.hub.broadcast <- msg
	}
}

// Client goroutine to write messages to client
func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close(103, "Disconnect")
	}()

	for {
		select {
		case message, ok := <-c.send:
			// c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				wsjson.Write(context.Background(), c.Conn, "Some issue")
				// c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			} else {
				err := wsjson.Write(context.Background(), c.Conn, message)
				if err != nil {
					fmt.Println("Error: ", err)
					break
				}
			}
		case <-ticker.C:
			// c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsjson.Write(context.Background(), c.Conn, "Ping message"); err != nil {
				return
			}
		}

	}
}

// Client closing channel to unregister client
func (c *Client) Close() {
	close(c.send)
}

// Function to handle websocket connection and register client to hub and start goroutines
func ServeWS(ctx *gin.Context, hub *Hub) {
	ws, err := websocket.Accept(ctx.Writer, ctx.Request, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	client := NewClient(ws, hub)

	hub.register <- client
	go client.Write()
	go client.Read()
}
