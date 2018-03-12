package core

// Group maintains the set of active clients and broadcasts messages to the
// clients.
type Group struct {
	//Group index
	Index int

	// Registered clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	RedisConn *RedisConnect
}

func NewGroup(index int,redisAddr string,redisChannel string,) *Group {
	redisConn := RedisConnect{}

	g := &Group{
		Index: index,
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		RedisConn:  &redisConn,
	}

	redisConn.InitConn(redisAddr)
	redisConn.Subscribe(redisChannel,g.OnMessage)

	return g
}

func (g *Group) Run(redisChannel string) {
	for {
		select {
		case client := <-g.Register:
			g.Clients[client] = true
		case client := <-g.unregister:
			if _, ok := g.Clients[client]; ok {
				delete(g.Clients, client)
				close(client.Send)
			}
		case message := <-g.broadcast:
			g.RedisConn.Publish(redisChannel,message)
		}
	}
}

//broadcast
func (g *Group)OnMessage(channel string, data []byte) error{
	for client := range g.Clients {
		if(client.Groups[g]){
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(g.Clients, client)
			}
		}
	}
	return nil
}
