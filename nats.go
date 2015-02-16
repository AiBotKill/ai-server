package main

import (
	"os"

	gnatsd "github.com/apcera/gnatsd/server"
	"github.com/apcera/nats"
)

var natsConn *nats.Conn
var natsEncodedConn *nats.EncodedConn

func debugGnatsd() {
	opts := gnatsd.Options{}
	s := gnatsd.New(&opts)
	go s.Start()
}

func startNats() error {
	addr := os.Getenv("GNATSD_PORT_4222_TCP_ADDR")
	port := os.Getenv("GNATSD_PORT_4222_TCP_PORT")
	url := ""
	if addr == "" || port == "" {
		url = "nats://localhost:4222"
	} else {
		url = ("nats://" + addr + ":" + port)
	}

	c, err := nats.Connect(url)
	if err != nil {
		return err
	}

	nc, err := nats.NewEncodedConn(c, "json")
	if err != nil {
		return err
	}
	natsConn = c
	natsEncodedConn = nc
	return nil
}
