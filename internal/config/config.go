package config

type Config struct {
	ServerCount int
}

// Option is an optional functionality for
// any speedtest client
type Option func(*Config)
