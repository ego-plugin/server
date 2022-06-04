package eref

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/gorilla/websocket"
	"net/http"
)

// BuildWebsocket ..
func BuildWebsocket(opts ...WebSocketOption) *WebSocket {
	return comp.BuildWebsocket(opts...)
}

// UpgradeFilter protocol to WebSocket
func UpgradeFilter(ws *WebSocket, handler WebSocketFunc) restful.FilterFunction {
	return Filter(func(ctx FilterContext) {
		comp.UpgradeRoute(ws, handler)
	})
}

// UpgradeRoute protocol to WebSocket
func UpgradeRoute(ws *WebSocket, handler WebSocketFunc) restful.RouteFunction {
	return comp.UpgradeRoute(ws, handler)
}

// UpgradeRoute protocol to WebSocket
func (c *Component) UpgradeRoute(ws *WebSocket, handler WebSocketFunc) restful.RouteFunction {
	return RouteContext(func(ctx Context) {
		ws.Upgrade(ctx.Resp(), ctx.Req(), ctx, handler)
	})
}

// BuildWebsocket ..
func (c *Component) BuildWebsocket(opts ...WebSocketOption) *WebSocket {
	upgrade := &websocket.Upgrader{}
	// 支持跨域
	if c.config.EnableWebsocketCheckOrigin {
		upgrade.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	ws := &WebSocket{
		Upgrader: upgrade,
	}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

type WebSocket struct {
	*websocket.Upgrader
}

// Upgrade get upgrage request
func (ws *WebSocket) Upgrade(w http.ResponseWriter, r *http.Request, ctx Context, handler WebSocketFunc) {
	// todo response Header
	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err == nil {
		defer conn.Close()
	}
	wsConn := &WebSocketConn{
		Conn: conn,
		Ctx:  ctx,
	}
	handler(wsConn, err)
}

// WebSocketFunc ..
type WebSocketFunc func(*WebSocketConn, error)

type WebSocketConn struct {
	*websocket.Conn
	Ctx Context
}
