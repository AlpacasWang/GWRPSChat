package core

import (
	"time"
	"fmt"
	"github.com/garyburd/redigo/redis"
)

var SubConn redis.PubSubConn
var PubConn redis.Conn
//死活監視間隔
const HealthCheckPeriod = time.Minute

func InitConn(address string) error{
	var err error
	var tmpConn redis.Conn 
	tmpConn, err = redis.Dial("tcp", address,
		// Read timeout on server should be greater than ping period.
		redis.DialReadTimeout(HealthCheckPeriod+10*time.Second),
		redis.DialWriteTimeout(10*time.Second))
	if err != nil {
		return err
	}
	SubConn = redis.PubSubConn{Conn: tmpConn}

	PubConn, err = redis.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func CloseConn(){
	SubConn.Unsubscribe()
	SubConn.Close()
	PubConn.Close()
}

func Publish(channel string,message []byte) {
	PubConn.Do("PUBLISH", channel, message)
}

// This example shows how receive pubsub notifications with cancelation and
// health checks.
func Subscribe(channel string,onMessage func(channel string, data []byte) error) {

	err := listenPubSubChannels(onMessage,"c1")

	if err != nil {
		fmt.Println(err)
		return
	}
}

// L listens for messages on Redis pubsub channels. The
// onStart function is called after the channels are subscribed. The onMessage
// function is called for each message.
func listenPubSubChannels(
	onMessage func(channel string, data []byte) error,
	channels ...string) error {
	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	if err := SubConn.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}
    fmt.Println("Subscribe Success ")
	done := make(chan error, 1)

	// Start a goroutine to receive notifications from the server.
	go func() {
		for {
			switch n := SubConn.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					done <- err
					return
				}
			case redis.Subscription:
				fmt.Println("Subscription Count Change:",n.Count) 
			}
		}
	}()

	//死活監視
	go func() error{
		var err error = nil
		ticker := time.NewTicker(HealthCheckPeriod)
		defer ticker.Stop()
		for err == nil {
			select {
			case <-ticker.C:
				// Send ping to test health of connection and server. If
				// corresponding pong is not received, then receive on the
				// connection will timeout and the receive goroutine will exit.
				if err = SubConn.Ping(""); err != nil {
					break 
				}
			case err := <-done:
				// Return error from the receive goroutine.
				return err
			}
		}
		return err
	}()

	// Wait for goroutine to complete.
	return nil
}

