package main

import (
	"log"
	"os"

	"github.com/getlantern/systray"
)

func main() {
	// In production/release, you might want to conditionally log to file
	// For open source release, we'll allow standard logging
	// Uncomment the following lines if you want to log to file:
	/*
	f, err := os.OpenFile("cliprelay.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(f)
		defer f.Close()
	}
	*/

	cfg, err := loadConfig()
	if err != nil {
		log.Printf("[config] error: %v — pakai default", err)
		cfg = defaultConfig
	}

	initState(cfg)
	systray.Run(onReady, func() {})
}