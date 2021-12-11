package eref

import (
	"bytes"
	"fmt"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// extractAPP 提取header头中的app信息
func extractAPP(req *restful.Request) string {
	return req.Request.Header.Get("app")
}

// recoverMiddleware 恢复拦截器，记录500信息，以及慢日志信息
func recoverMiddleware(logger *elog.Component, config *Config) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var beg = time.Now()
		// 为了性能考虑，如果要加日志字段，需要改变slice大小
		var fields = make([]elog.Field, 0, 15)
		var brokenPipe bool
		var event = "normal"
		defer func() {
			cost := time.Since(beg)

			fields = append(fields,
				elog.FieldCost(cost),
				elog.FieldType(req.Request.Method),     // GET, POST
				elog.FieldMethod(req.Request.URL.Path), // 完整路径
				elog.FieldAddr(req.Request.URL.Path),
				elog.FieldIP(clientIP(req)),
				elog.FieldSize(int32(resp.ContentLength())),
				elog.FieldPeerIP(getPeerIP(req.Request.RemoteAddr)),
			)

			if config.EnableTraceInterceptor && opentracing.IsGlobalTracerRegistered() {
				fields = append(fields, elog.FieldTid(etrace.ExtractTraceID(req.Request.Context())))
			}

			// slow log
			if config.SlowLogThreshold > time.Duration(0) && config.SlowLogThreshold < cost {
				logger.Warn("slow", fields...)
			}

			if rec := recover(); rec != nil {
				if ne, ok := rec.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					_ = resp.WriteError(http.StatusInternalServerError, rec.(error)) // nolint: errcheck
				} else {
					resp.WriteHeader(http.StatusInternalServerError)
				}

				event = "recover"
				stackInfo := stack(3)

				fields = append(fields,
					elog.FieldEvent(event),
					zap.ByteString("stack", stackInfo),
					elog.FieldErrAny(rec),
					elog.FieldCode(int32(resp.StatusCode())),
					elog.FieldUniformCode(int32(resp.StatusCode())),
				)
				logger.Error("access", fields...)
				return
			}

			fields = append(fields,
				elog.FieldEvent(event),
				elog.FieldErrAny(""),
				elog.FieldCode(int32(resp.StatusCode())),
			)
			logger.Info("access", fields...)
		}()
		chain.ProcessFilter(req, resp)
	}
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func metricServerInterceptor() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		beg := time.Now()
		chain.ProcessFilter(req, resp)
		emetric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), emetric.TypeHTTP, req.Request.Method+"."+req.Request.URL.Path, extractAPP(req))
		emetric.ServerHandleCounter.Inc(emetric.TypeHTTP, req.Request.Method+"."+req.Request.URL.Path, extractAPP(req), http.StatusText(resp.StatusCode()), http.StatusText(resp.StatusCode()))
	}
}

func traceServerInterceptor() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		span, ctx := etrace.StartSpanFromContext(
			req.Request.Context(),
			req.Request.Method+"."+req.Request.URL.Path,
			etrace.TagComponent("http"),
			etrace.TagSpanKind("server"),
			etrace.HeaderExtractor(req.Request.Header),
			etrace.CustomTag("http.url", req.Request.URL.Path),
			etrace.CustomTag("http.method", req.Request.Method),
			etrace.CustomTag("peer.ipv4", clientIP(req)),
		)
		req.Request = req.Request.WithContext(ctx)
		defer span.Finish()
		// 判断了全局jaeger的设置，所以这里一定能够断言为jaeger
		resp.AddHeader(eapp.EgoTraceIDName(), span.(*jaeger.Span).Context().(jaeger.SpanContext).TraceID().String())
		chain.ProcessFilter(req, resp)
	}
}

// IP returns the IP address of request.
func clientIP(req *restful.Request) string {
	var ip string
	ra := req.Request.RemoteAddr
	if ip = req.HeaderParameter("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip = req.HeaderParameter("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// 获取对端ip
func getPeerIP(addr string) string {
	addSlice := strings.Split(addr, ":")
	if len(addSlice) > 1 {
		return addSlice[0]
	}
	return ""
}
