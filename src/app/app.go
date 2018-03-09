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

const addr = ":8080"
const redisAddr = "127.0.0.1:6379"
const defaultChannel = "c1"
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
	core.InitConn(redisAddr);
	core.Subscribe(defaultChannel,hub.OnMessage)
	go hub.Run(defaultChannel)
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs( w, r)
	})
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}