package config

type Config struct {
	ServerCount int
	Token       string
}

// Option is an optional functionality for
// any speedtest client
type Option func(*Config)
