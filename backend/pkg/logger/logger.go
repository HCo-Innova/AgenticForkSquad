package logger

import "log"

func Info(msg string) {
	log.Printf("[INFO] %s", msg)
}

func Error(msg string, err error) {
	log.Printf("[ERROR] %s: %v", msg, err)
}

func Debug(msg string) {
	log.Printf("[DEBUG] %s", msg)
}
