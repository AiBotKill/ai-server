package main

import (
	"log"
	"time"
)

var serviceId string

func main() {
	serviceId = Uuid()
	if err := startNats(); err != nil {
		log.Panicln("Can connect or start to gnatsd:", err.Error())
	}

	sockServer := NewSocketServer()
	go sockServer.Listen(2000)

	websockServer := NewWebsocketServer()
	go websockServer.Listen(2001)

	// Logging server stats until the server is stopped.
	for {
		<-time.After(time.Second * 5)
		err := natsEncodedConn.Publish("ping", map[string]interface{}{
			"ping":      "aiServer",
			"serviceId": serviceId,
			"time":      time.Now().String(),
		})
		if err != nil {
			log.Println(err.Error())
		}
	}
}
