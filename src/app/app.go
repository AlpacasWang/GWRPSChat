package app

import (
	"log"
	"fmt"
	"net/http"
	"core"
	"github.com/gorilla/websocket"
)

const addr = ":8080"
const redisAddr = "127.0.0.1:6379"
const defaultChannel = "c1"
const defaultGroup = 0
var groups = make(map[int]*core.Group)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	g, ok := groups[defaultGroup]
	if(!ok){
		g = core.NewGroup(defaultGroup,redisAddr,defaultChannel)
		groups[defaultGroup] = g
		go g.Run(defaultChannel)
	}
	client := &core.Client{Groups: map[*core.Group]bool{g:true},Conn: conn, Send: make(chan []byte, 256)}
	g.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}

func Run() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", ServeWs)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}



