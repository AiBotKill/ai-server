package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type SocketServer struct {
	Id string
}

func NewSocketServer() *SocketServer {
	s := &SocketServer{}
	s.Id = Uuid()
	return s
}

func (s *SocketServer) Listen(port int) {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println("SocketServer.Listen Error:", err.Error())
		return
	}
	defer l.Close()
	for {
		sock, err := l.Accept()
		LogDebug("New connection from", l.Addr())
		if err != nil {
			log.Println("SocketServer.Listen.Accept Error:", err.Error())
		}

		ac, err := NewSockAiConn(sock)
		if err != nil {
			log.Println("Can't accept connection:", err)
			continue
		}
		aiConn := NewAiConn(ac)
		go aiConn.Parser()
	}
}

type SockAiConn struct {
	Conn      net.Conn
	WriteLock *sync.Mutex
	Scanner   *bufio.Scanner
	Inbound   chan string
	Outbound  chan string
}

func NewSockAiConn(conn net.Conn) (*SockAiConn, error) {
	c := &SockAiConn{
		Conn:      conn,
		WriteLock: &sync.Mutex{},
		Scanner:   bufio.NewScanner(conn),
		Inbound:   make(chan string),
		Outbound:  make(chan string),
	}
	go c.reader()
	return c, nil
}

func (c *SockAiConn) Read() (string, error) {
	LogDebug("reading...")
	msg, ok := <-c.Inbound
	LogDebug("msg")
	if !ok {
		return "", errors.New("Error while reading from channel.")
	}
	return string(msg), nil
}

func (s *SockAiConn) Write(line string) error {
	return s.write([]byte(line))
}

func (s *SockAiConn) Close() error {
	return s.Conn.Close()
}

// write will write []byte to AiConn.sc with lock, so we don't get corrupted messages with overlapping writes
func (s *SockAiConn) write(b []byte) error {
	s.WriteLock.Lock()
	if _, err := s.Conn.Write(b); err != nil {
		return err
	}
	if _, err := s.Conn.Write([]byte("\n")); err != nil {
		return err
	}
	s.WriteLock.Unlock()
	return nil
}

func (c *SockAiConn) reader() {
	for {
		s, err := bufio.NewReader(c.Conn).ReadString('\n')
		// Remove newlines and carriage returns.
		s = strings.Replace(s, "\r", "", -1)
		s = strings.Replace(s, "\n", "", -1)
		if err != nil {
			c.Conn.Close()
			LogError("Socket reader stopping:", err.Error())
			close(c.Inbound)
			return
		}
		c.Inbound <- s
	}
}
