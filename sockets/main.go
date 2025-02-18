package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrades connection from http to ws
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

var (
	clients    = make(map[*Client]bool)
	broadcast  = make(chan []byte)
	register   = make(chan *Client)
	deregister = make(chan *Client)
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer conn.Close()

	client := &Client{conn: conn, send: make(chan []byte)}

	register <- client

	go client.readMessages()
	client.writeMessages()

}

func (c *Client) readMessages() {
	defer func() {
		deregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("dude error reading message:,", err)
			break
		}

		broadcast <- message
	}

}

func (c *Client) writeMessages() {
	defer c.conn.Close()

	for {
		message, ok := <-c.send
		if !ok {
			// If the channel is closed, it means the client is disconnected
			return
		}
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

func handleMessages() {
	for {
		select {
		case client := <-register:
			clients[client] = true
		case client := <-deregister:
			if _, ok := clients[client]; ok {
				delete(clients, client)
				close(client.send)
				log.Println("Client disconnected")
			}
		case message := <-broadcast:
			for client := range clients {
				select {
				case client.send <- message:
				default:
					delete(clients, client)
					close(client.send)
				}
			}
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("/Users/allenkimanideep/hld/sockets/static/"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Chat server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
