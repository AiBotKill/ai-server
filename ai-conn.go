package main

import (
	"encoding/json"
	"time"
)

const NATS_TIMEOUT = time.Second * 3

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
	BotIds []string
}

func NewAiConn(conn ReadWriteCloser) *AiConn {
	a := &AiConn{}
	a.Conn = conn
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
				return
			}
			LogDebug("got string", line)

			// If teamId has not been set yet, expect the first message
			// to be register message. Keep trying until registration
			// is acknowledged by ai-console.
			if a.TeamId == "" {
				var regMsg RegisterMessage
				err := json.Unmarshal([]byte(line), &regMsg)
				if err != nil {
					e := "Registration message unmarshal error: " + err.Error()
					LogWarn(e)
					a.Conn.Write(NewJsonError(e))
					continue
				}

				// Request registration through NATS
				var regReply RegisterReply
				err = natsEncodedConn.Request("registerAi", regMsg, &regReply, NATS_TIMEOUT)
				if err != nil {
					e := "Registration message publish error: " + err.Error()
					LogError(e)
					a.Conn.Write(NewJsonError(e))
					continue
				}

				// Check that reply status is ok.
				if regReply.Status != "ok" {
					e := "Registration not ok for team " + regMsg.TeamId + ": " + regReply.Error
					LogWarn(e)
					a.Conn.Write(line)
					continue
				}

				// All ok, Set TeamId
				a.TeamId = regReply.Id
				a.Conn.Write(line)
				continue
			}
		}
	}()
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
	TeamId  string `json:"botId"`
	Version string `json:"version"`
}

type RegisterReply struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	Error  string `json:"error,omitempty"`
}
