package config

import (
	"fmt"
)

type ServerConfig struct {
	IP   string
	Port int
}

func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.IP, s.Port)
}

// LoadConfig returns a list of servers
func LoadConfig() []*ServerConfig {
	return []*ServerConfig{
		{IP: "127.0.0.1", Port: 8081},
		{IP: "127.0.0.1", Port: 8082},
		{IP: "127.0.0.1", Port: 8083},
	}
}
