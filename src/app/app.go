// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"flag"
	"log"
	"net/http"
	"core"
)

var addr = flag.String("addr", ":8080", "http service address")
var hub *core.Hub
func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func Run() {
	flag.Parse()
	hub = core.NewHub()
	core.InitConn("127.0.0.1:6379");
	core.Subscribe("c1",onMessage)
	go hub.Run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs( w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func onMessage(channel string, data []byte) error{
	for client := range hub.Clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(hub.Clients, client)
		}
	}
	return nil
}