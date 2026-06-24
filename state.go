package main

import (
	"sync"
	"sync/atomic"
)

var (
	stateMu    sync.Mutex
	buf        string
	activeCfg  Config
	isTyping   atomic.Bool
	isFetching atomic.Bool
)

func initState(cfg Config) {
	stateMu.Lock()
	activeCfg = cfg
	stateMu.Unlock()
}

func getBuffer() string {
	stateMu.Lock()
	defer stateMu.Unlock()
	return buf
}

func setBuffer(s string) {
	stateMu.Lock()
	buf = s
	stateMu.Unlock()
}

func getCfg() Config {
	stateMu.Lock()
	defer stateMu.Unlock()
	return activeCfg
}

func setCfg(cfg Config) {
	stateMu.Lock()
	activeCfg = cfg
	stateMu.Unlock()
}
