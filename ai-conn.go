package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/apcera/nats"
)

const NATS_TIMEOUT = time.Second * 10

type Reader interface {
	Read() (line string, err error)
}

type Writer interface {
	Write(line string) error
}

type Closer interface {
	Close() error
}

type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

type AiConn struct {
	Conn   ReadWriteCloser
	TeamId string
	BotId  string
	GameId string
	State  string
}

func NewAiConn(conn ReadWriteCloser) *AiConn {
	a := &AiConn{}
	a.Conn = conn
	a.State = "unregistered"
	return a
}

func (a *AiConn) Parser() {
	go func() {
		for {
			line, err := a.Conn.Read()
			// If error happens when reading from ai-connection, we will close the connection.
			if err != nil {
				LogError("Error while reading from ai connection, closing: " + err.Error())
				a.Conn.Close()
				if a.State == "registered" || a.State == "joined" {
					err = natsEncodedConn.Publish("unregisterAI", map[string]interface{}{
						"botId": a.BotId,
					})
					if err != nil {
						LogError("Can't unregister ai:", err.Error())
					}
				}
				return
			}
			LogDebug("got string", line)

			// If teamId has not been set yet, expect the first message
			// to be register message. Keep trying until registration
			// is acknowledged by ai-console.
			if a.State == "unregistered" {
				var regMsg RegisterMessage
				err := json.Unmarshal([]byte(line), &regMsg)
				if err != nil {
					a.LogErr(err)
					continue
				}

				// First message is ALWAYS registerAi, so we
				// help out a bit.
				regMsg.Type = "registerAi"

				// Request registration through NATS
				var regReply Reply
				err = natsEncodedConn.Request("registerAI", regMsg, &regReply, NATS_TIMEOUT)
				if err != nil {
					a.LogErr(err)
					continue
				}

				// DEBUG
				if b, err := json.Marshal(&regReply); err == nil {
					LogDebug("Got reply!", string(b))
				}

				// Check that reply status is ok.
				if regReply.Status != "ok" {
					err := errors.New("Registration not ok for team " + regMsg.TeamId + ": " + regReply.Error)
					a.LogErr(err)
					continue
				}

				// All ok, Set TeamId
				a.BotId = regReply.Id
				b, _ := json.Marshal(regReply)
				a.Conn.Write(string(b))
				a.State = "joined"

				natsConn.Subscribe(a.BotId+".gameState", func(msg *nats.Msg) {
					err := a.Conn.Write(string(msg.Data))
					if err != nil {
						a.LogErr(err)
					}
					natsConn.Publish(msg.Reply, NewReply(a.BotId, err))
				})
				continue
			}

			if a.State == "joined" {
				var action ActionRequest
				err := json.Unmarshal([]byte(line), &action)
				if err != nil {
					a.LogErr(err)
					continue
				}

				log.Println("got action ok", action)

				var reply Reply
				if err := natsEncodedConn.Request(a.BotId+".action", action, &reply, NATS_TIMEOUT); err != nil {
					a.LogErr(err)
					continue
				} else {
					log.Println("sent action ok")
				}

				replBytes, err := json.Marshal(&reply)
				if err != nil {
					a.LogErr(err)
					continue
				}

				a.Conn.Write(string(replBytes))
			}
		}
	}()
}

func NewReply(id string, err error) []byte {
	reply := &Reply{
		Type: "reply",
		Id:   id,
	}
	if err != nil {
		reply.Status = "error"
		reply.Error = err.Error()
	} else {
		reply.Status = "ok"
	}

	b, _ := json.Marshal(&reply)
	return b
}

func (c *AiConn) LogErr(err error) {
	LogError(err.Error())
	b, _ := json.Marshal(map[string]string{
		"type":   "reply",
		"status": "error",
		"error":  err.Error(),
	})
	c.Conn.Write(string(b))
}

func GetMsgType(b []byte) string {
	var m struct {
		Type string `json:"type"`
	}
	err := json.Unmarshal(b, m)
	if err != nil {
		return ""
	}
	return m.Type
}

func NewJsonError(e string) string {
	v := map[string]string{
		"status": "error",
		"error":  e,
	}
	b, err := json.Marshal(&v)
	if err != nil {
		return ""
	}
	return string(b)
}

type RegisterMessage struct {
	Type    string `json:"type"`
	TeamId  string `json:"teamId"`
	Version string `json:"version"`
}

type Reply struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Id     string `json:"id"`
	Error  string `json:"error,omitempty"`
}

type JoinRequest struct {
	Type     string `json:"type"`
	GameId   string `json:"gameId"`
	GameMode string `json:"gameMode"`
}

type GameState struct {
	Type string `json:"type"`
}

type ActionRequest struct {
	Type      string `json:"type"`
	Direction struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"direction"`
}

type JoinGame struct {
	Type   string `json:"joinGame"`
	BotId  string `json:"botid"`
	GameId string `jsno:"gameId"`
}
