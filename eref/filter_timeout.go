package eref

import (
	"context"
	"github.com/emicklei/go-restful/v3"
	"net/http"
	"time"
)

// timeout middleware wraps the request context with a timeout
func timeoutMiddleware(timeout time.Duration) restful.FilterFunction {
	return Filter(func(c FilterContext) {
		// 若无自定义超时设置，默认设置超时
		_, ok := c.Req().Context().Deadline()
		if ok {
			c.ProcessFilter()
			return
		}

		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Req().Context(), timeout)
		defer func() {
			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {
				// write response and abort the request
				c.Response.WriteHeader(http.StatusGatewayTimeout)
				c.FilterChain.Index = 63
			}
			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		c.Request.Request = c.Req().WithContext(ctx)
		c.ProcessFilter()
	})
}
