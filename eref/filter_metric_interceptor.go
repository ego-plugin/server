package eref

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/emetric"
	"net/http"
	"time"
)

func metricServerInterceptor() restful.FilterFunction {
	return Filter(func(ctx FilterContext) {
		beg := time.Now()
		ctx.ProcessFilter()
		emetric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), emetric.TypeHTTP, ctx.Req().Method+"."+ctx.SelectedRoutePath(), extractAPP(ctx.Request))
		emetric.ServerHandleCounter.Inc(emetric.TypeHTTP, ctx.Req().Method+"."+ctx.SelectedRoutePath(), extractAPP(ctx.Request), http.StatusText(ctx.StatusCode()), http.StatusText(ctx.StatusCode()))
	})
}
