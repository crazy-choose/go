package ws

import "time"

type Config struct {
	WriteWait         time.Duration
	PongWait          time.Duration
	PingPeriod        time.Duration
	MaxMessageSize    int64
	MessageBufferSize int
}

func newConfig() *Config {
	return &Config{
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        10 * time.Second,
		MaxMessageSize:    512,
		MessageBufferSize: 256,
	}
}
