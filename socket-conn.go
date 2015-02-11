package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

// AiConn is a struct for one ai-connection through socket.
type AiConn struct {
	Id        string      `json:"id"`
	sc        net.Conn    `json:"-"`
	writeLock *sync.Mutex `json:"-"`
	botId     string      `json:"botId"`
	version   string      `json"version"`
}

// NewAiConn will return a new AiConn, takes a EncodedConn as parameter,
func NewAiConn() (*AiConn, error) {
	ac := &AiConn{}
	ac.Id = Uuid()
	ac.writeLock = &sync.Mutex{}
	return ac, nil
}

// HandleConnection will handle socket connections and messages coming in and out.
func (ac *AiConn) HandleConnection(sc net.Conn) {
	natsEncodedConn.Publish("log", map[string]interface{}{
		"msg": "AiConn '" + ac.Id + "' connection handler starting.",
	})
	log.Println("AiConn '" + ac.Id + "' connection handler starting.")
	ac.sc = sc
	scanner := bufio.NewScanner(ac.sc)
	for scanner.Scan() {
		// Read line
		line := scanner.Bytes()
		log.Println("AiConnection \""+ac.Id+"\" got line:", string(line))

		// if botid is not set, the ai has not registered. check if this
		// is the register message, then post it to console and wait
		// for answer. if console think's it's ok, set botid and reply
		// to ai with success.
		if ac.botId == "" {
			var m RegisterMessage
			err := json.Unmarshal(line, &m)
			if err != nil {
				ac.logErr(err)
				continue
			}

			var resp IdReplyMsg
			err = natsEncodedConn.Request("registerAI", m, &resp, 3*time.Second)
			if err != nil {
				ac.logErr(err)
			} else {
				log.Println(resp)
				// All ok, marshal response,
				b, err := json.Marshal(&resp)
				if err != nil {
					ac.logErr(err)
				}
				ac.write(b)
				ac.botId = m.BotId
				ac.version = m.Version
			}
			continue
		}
	}
	natsEncodedConn.Publish("log", map[string]interface{}{
		"msg": "AiConn '" + ac.Id + "' connection handler stopped.",
	})
	log.Println("AiConn '" + ac.Id + "' connection handler stopped.")
}

func (ac *AiConn) logErr(err error) {
	log.Println(err.Error())
	airepl := &IdReplyMsg{
		Status: "error",
		Id:     ac.botId,
		Error:  err.Error(),
	}
	b, err := json.Marshal(&airepl)
	if err != nil {
		log.Println(err.Error())
		return
	}
	ac.write(b)
}

// write will write []byte to AiConn.sc with lock, so we don't get corrupted messages with overlapping writes
func (ac *AiConn) write(b []byte) error {
	ac.writeLock.Lock()
	if _, err := ac.sc.Write(b); err != nil {
		return err
	}
	if _, err := ac.sc.Write([]byte("\n")); err != nil {
		return err
	}
	ac.writeLock.Unlock()
	return nil
}

type IdReplyMsg struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	Error  string `json:"error,omitempty"`
}

type RegisterMessage struct {
	BotId   string `json:"botId"`
	Version string `json:"version"`
}
