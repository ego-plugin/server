package erestful

import (
	restful "github.com/emicklei/go-restful/v3"
	"github.com/gorilla/websocket"
	"net/http"
)

// Upgrade protocol to WebSocket
func (c *Component) Upgrade(rb *restful.RouteBuilder, ws *WebSocket, handler WebSocketFunc) *restful.RouteBuilder {
	rb.To(func(req *restful.Request, resp *restful.Response) {
		ws.Upgrade(req, resp, handler)
	})
	return rb
}

// BuildWebsocket ..
func (c *Component) BuildWebsocket(opts ...WebSocketOption) *WebSocket {
	upgrader := &websocket.Upgrader{}
	// 支持跨域
	if c.config.EnableWebsocketCheckOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	ws := &WebSocket{
		Upgrader: upgrader,
	}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

// WebSocket ..
type WebSocket struct {
	*websocket.Upgrader
}

// Upgrade get upgrage request
func (ws *WebSocket) Upgrade(req *restful.Request, resp *restful.Response, handler WebSocketFunc) {
	// todo response Header
	conn, err := ws.Upgrader.Upgrade(resp.ResponseWriter, req.Request, resp.Header())
	if err == nil {
		defer conn.Close()
	}
	wsConn := &WebSocketConn{
		Conn:   conn,
		Request: req,
		Response: resp,
	}
	handler(wsConn, err)
}

// WebSocketFunc ..
type WebSocketFunc func(*WebSocketConn, error)

// WebSocketConn ...
type WebSocketConn struct {
	*websocket.Conn
	*restful.Request
	*restful.Response
}