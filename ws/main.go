package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}

type Client struct {
	world *World
	conn  *websocket.Conn
	send  chan []byte
}

type World struct {
	clientMap map[*Client]bool
	ChanEnter chan *Client
	ChanLeave chan *Client
	broadcast chan []byte
}

func (w *World) run() {
	w.ChanEnter = make(chan *Client)
	w.ChanLeave = make(chan *Client)

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case client := <-w.ChanEnter:
			log.Println("Client Entered")
			w.clientMap[client] = true
		case client := <-w.ChanLeave:
			log.Println("Client Leave")
			if _, ok := w.clientMap[client]; ok {
				delete(w.clientMap, client)
				close(client.send)
			}
		case message := <-w.broadcast:
			for client := range w.clientMap {
				client.send <- message
			}
		case tick := <-ticker.C:
			for c := range w.clientMap {
				c.send <- []byte(tick.String())
			}
		}

	}
}

func newWorld() *World {
	return &World{
		clientMap: make(map[*Client]bool, 5),
		broadcast: make(chan []byte),
	}
}

func NewClient(w *World, c *websocket.Conn) (client *Client) {
	client = &Client{
		world: w,
		conn:  c,
		send:  make(chan []byte, 256),
	}

	go client.readPump()
	go client.writePump()
	return client
}

func (c *Client) readPump() {
	defer func() {
		c.world.ChanLeave <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			log.Println(err)
			break
		}
		c.world.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

}

func main() {
	r := gin.Default()

	world := newWorld()
	go world.run()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		cl := NewClient(world, conn)
		cl.world.ChanEnter <- cl
	})

	r.Run()

}
