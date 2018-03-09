// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import(
	"fmt"
	"net/http"
)
var DefaultGroups = map[int]bool{0:true}
const DefaultGroup = 0
// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run(redisChannel string) {
	for {
		select {
		case client := <-h.register:
			h.Clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.broadcast:
			Publish(redisChannel,message)
		}
	}
}

//broadcast
func (hub *Hub)OnMessage(channel string, data []byte) error{
	for client := range hub.Clients {
		if(client.Groups[DefaultGroup]){
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(hub.Clients, client)
			}
		}
	}
	return nil
}


// serveWs handles websocket requests from the peer.
func (hub *Hub)ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &Client{hub: hub, Groups: DefaultGroups,conn: conn, Send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
