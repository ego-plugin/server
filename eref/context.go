package eref

import (
	"context"
	"github.com/ego-plugin/binding"
	"github.com/emicklei/go-restful/v3"
	"net"
	"net/http"
)

type Context struct {
	*restful.Request
	*restful.Response
}

func (c Context) BindQuery(v any) error {
	return binding.Query.Bind(c.Request.Request, v, binding.LANG_EN)
}

func (c Context) Bind(v any) error {
	return binding.Form.Bind(c.Request.Request, v, binding.LANG_EN)
}

func (c Context) BindMsgPack(v any) error {
	return binding.MsgPack.Bind(c.Request.Request, v, binding.LANG_EN)
}

func (c Context) ClientIP() string {
	if ip, ok := c.Request.Attribute("ip").(string); ok {
		return ip
	}
	return getIP(c.Req())
}

func (c Context) GetPeerIP() string {
	addr, _, err := net.SplitHostPort(c.Req().RemoteAddr)
	if err != nil {
		return ""
	}
	return addr
}

func (c Context) Context() context.Context {
	return c.Request.Request.Context()
}

func (c Context) Req() *http.Request {
	return c.Request.Request
}

func (c Context) Resp() http.ResponseWriter {
	return c.Response.ResponseWriter
}

func (c Context) BodyToByte() []byte {
	if b, o := c.Request.Attribute("body").([]byte); o {
		return b
	}
	return []byte("")
}

type RouteContextFunc func(ctx Context)

func RouteContext(f RouteContextFunc) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		c := Context{
			Request:  req,
			Response: resp,
		}
		f(c)
	}
}

func NewRoute(prefixUrl string) *restful.WebService {
	ws := new(restful.WebService)
	return ws.Path(prefixUrl).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
}

// FilterContext 中间件上下文
type FilterContext struct {
	Context
	*restful.FilterChain
}

func (c *FilterContext) ProcessFilter() {
	c.FilterChain.ProcessFilter(c.Request, c.Response)
}

type FilterContextFunc func(ctx FilterContext)

func Filter(f FilterContextFunc) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		c := FilterContext{
			Context: Context{
				Request:  req,
				Response: resp,
			},
			FilterChain: chain,
		}
		f(c)
	}
}
