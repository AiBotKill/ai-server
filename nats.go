package main

import (
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
	c, err := nats.Connect(nats.DefaultURL)
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
