package main

import "log"
import "fmt"

func LogFatal(v ...interface{}) {
	natsConn.Publish("log.fatal", []byte(fmt.Sprintln(v...)))
	log.Println("[ERROR]", v)
}

func LogError(v ...interface{}) {
	natsConn.Publish("log.error", []byte(fmt.Sprintln(v...)))
	log.Println("[ERROR]", v)
}

func LogWarn(v ...interface{}) {
	natsConn.Publish("log.warn", []byte(fmt.Sprintln(v...)))
	log.Println("[WARN]", v)
}

func LogInfo(v ...interface{}) {
	natsConn.Publish("log.info", []byte(fmt.Sprintln(v...)))
	log.Println("[INFO]", v)
}

func LogDebug(v ...interface{}) {
	natsConn.Publish("log.debug", []byte(fmt.Sprintln(v...)))
	log.Println("[DEBUG]", v)
}
