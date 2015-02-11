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

	sc := NewAiServer()
	go sc.Listen(2000)

	// Logging server stats until the server is stopped.
	for {
		<-time.After(time.Second * 1)
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
