package erestful

// Option 可选项
type Option func(c *Container)

// WebSocketOption ..
type WebSocketOption func(*WebSocket)