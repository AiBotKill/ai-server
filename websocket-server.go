package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	readWait   = 10 * time.Second
	pingPeriod = (readWait * 9) / 10
)

type WebsocketServer struct {
	Router *mux.Router `json:"router"`
}

func NewWebsocketServer() *WebsocketServer {
	s := &WebsocketServer{}
	s.Router = mux.NewRouter()
	s.Router.PathPrefix("/").HandlerFunc(s.websocketHandler)
	http.Handle("/", s.Router)
	return s
}

func (s *WebsocketServer) Listen(port int) {
	LogDebug("Listen...")
	go http.ListenAndServe(":"+strconv.Itoa(port), nil)
	LogDebug("Listen...")
}
func (s *WebsocketServer) websocketHandler(w http.ResponseWriter, r *http.Request) {
	LogDebug("new connection")
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			http.Error(w, "Not a websocket handshake", 400)
		}
		LogError("Error while upgrading:", err.Error())
		return
	}

	conn := NewWebsocketConn(ws)
	go conn.reader()
	go conn.writer()
	ac := NewAiConn(conn)
	go ac.Parser()
}

type WebsocketConn struct {
	Conn     *websocket.Conn
	Inbound  chan []byte
	Outbound chan []byte
	Closed   bool
}

func NewWebsocketConn(conn *websocket.Conn) *WebsocketConn {
	c := &WebsocketConn{
		Conn:     conn,
		Inbound:  make(chan []byte),
		Outbound: make(chan []byte),
		Closed:   false,
	}
	return c
}

func (c *WebsocketConn) Write(line string) error {
	if c.Closed {
		return errors.New("Connection already closed")
	}
	c.Outbound <- []byte(line)
	return nil
}

func (c *WebsocketConn) Read() (string, error) {
	msg, ok := <-c.Inbound
	if !ok {
		return "", errors.New("Error while reading from channel.")
	}
	return string(msg), nil
}

func (c *WebsocketConn) Close() error {
	return nil
}

// reader is started as a routine, it will continue to read data from
// websocket connection and sends it to the connections inbound channel
// as strings
func (c *WebsocketConn) reader() {
	LogDebug("connection reader gorouting starting.")
	defer func() {
		LogDebug("connection reader gorouting stopping.")
		close(c.Inbound)
		c.Conn.Close()
	}()
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(writeWait))
		return nil
	})
	c.Conn.SetReadDeadline(time.Now().Add(readWait))
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			LogError("Reader error:", err.Error())
			break
		}
		//LogDebug("gotmessage:", string(message))
		c.Inbound <- message
	}
}

// Write message as byte array to connection, with messagetype
func (c *WebsocketConn) write(mt int, payload []byte) error {
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Conn.WriteMessage(mt, payload)
}

// Routine to continue to write from outbound channel to websocket
// connection. Will close outbound channel when closed.
func (c *WebsocketConn) writer() {
	LogDebug("connection writer gorouting starting.")
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		LogDebug("connection writer gorouting stopping.")
		c.Closed = true
		pingTicker.Stop()
		close(c.Outbound)
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Outbound:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				LogError("[connection.writePump] !ok.")
				return
			}
			if err := c.write(websocket.TextMessage, []byte(message)); err != nil {
				LogError("[connection.writePump] err: '", err, "'.")
				return
			}
		// When pingTicker ticks, send a PingMessage to client.
		case <-pingTicker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				LogError("[connection.writePump] pingTicker err: '", err, "'.")
				return
			}
		}
	}
}
