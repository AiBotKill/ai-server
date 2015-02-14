package main

import "log"
import "fmt"

func LogError(v ...interface{}) {
	log.Println("[ERROR]", v)
	LogNats("[ERROR]", v)
}

func LogWarn(v ...interface{}) {
	log.Println("[WARN]", v)
	LogNats("[WARN]", v)
}

func LogInfo(v ...interface{}) {
	log.Println("[INFO]", v)
	LogNats("[INFO]", v)
}

func LogDebug(v ...interface{}) {
	log.Println("[DEBUG]", v)
	LogNats("[DEBUG]", v)
}

func LogNats(v ...interface{}) {
	_ = natsEncodedConn.Publish("log", fmt.Sprintln(v))
}
