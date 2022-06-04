package eref

import "time"

// Option 可选项
type Option func(c *Container)

// WebSocketOption ..
type WebSocketOption func(*WebSocket)

// WithHost 设置host
func WithHost(host string) Option {
	return func(c *Container) {
		c.config.Host = host
	}
}

// WithPort 设置port
func WithPort(port int) Option {
	return func(c *Container) {
		c.config.Port = port
	}
}

// WithNetwork 设置network
func WithNetwork(network string) Option {
	return func(c *Container) {
		c.config.Network = network
	}
}

// WithServerReadTimeout 设置超时时间
func WithServerReadTimeout(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.ServerReadTimeout = timeout
	}
}

// WithServerReadHeaderTimeout 设置超时时间
func WithServerReadHeaderTimeout(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.ServerReadHeaderTimeout = timeout
	}
}

// WithServerWriteTimeout 设置超时时间
func WithServerWriteTimeout(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.ServerWriteTimeout = timeout
	}
}

// WithContextTimeout 设置port
func WithContextTimeout(timeout time.Duration) Option {
	return func(c *Container) {
		c.config.ContextTimeout = timeout
	}
}
