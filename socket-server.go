package main

import (
	"log"
	"net"
)

type SocketServer struct {
	Id      string    `json:"id"`
	AiConns []*AiConn `json:"scConns"`
}

func NewAiServer() *SocketServer {
	sc := &SocketServer{}
	sc.Id = Uuid()
	return sc
}

func (sc *SocketServer) Listen(port int) {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Println("SocketServer.Listen Error:", err.Error())
		return
	}
	defer l.Close()
	for {
		sock, err := l.Accept()
		if err != nil {
			log.Println("SocketServer.Listen.Accept Error:", err.Error())
		}
		ac, err := NewAiConn()
		if err != nil {
			log.Println("Can't accept connection:", err)
			continue
		}
		go ac.HandleConnection(sock)
		sc.AddAiConn(ac)
	}
}

func (sc *SocketServer) AddAiConn(ac *AiConn) {
	log.Println("Adding sc connection")
	sc.AiConns = append(sc.AiConns, ac)
}

func (sc *SocketServer) RmAiConn(ac *AiConn) {
	log.Println("Removing sc connection")
	for i := range sc.AiConns {
		if sc.AiConns[i] == ac {
			sc.AiConns = append(sc.AiConns[:i], sc.AiConns[i+1:]...)
			return
		}
	}
}
