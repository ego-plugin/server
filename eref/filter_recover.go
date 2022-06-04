package eref

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"io"
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

type resWriter struct {
	restful.EntityReaderWriter
	body *bytes.Buffer
}

func (w *resWriter) Write(resp *restful.Response, status int, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err = w.body.Write(b); err != nil {
		return err
	}
	return w.EntityReaderWriter.Write(resp, status, v)
}

// extractAPP 提取header头中的app信息
func extractAPP(req *restful.Request) string {
	return req.Request.Header.Get("app")
}

// recoverMiddleware 恢复拦截器，记录500信息，以及慢日志信息
func recoverMiddleware(logger *elog.Component, config *Config) restful.FilterFunction {
	return Filter(func(ctx FilterContext) {
		var rb bytes.Buffer
		var rw *resWriter
		var ok bool
		var entity restful.EntityReaderWriter

		config.mu.RLock()
		// 保存body
		ctx.Req().Body = ioutil.NopCloser(io.TeeReader(ctx.Req().Body, &rb))
		ctx.SetAttribute("body", rb.Bytes())

		if config.EnableAccessInterceptorRes {
			if entity, ok = ctx.Response.EntityWriter(); ok {
				rw = &resWriter{
					EntityReaderWriter: entity,
					body:               new(bytes.Buffer),
				}
			}
		}
		config.mu.RUnlock()

		var beg = time.Now()
		// 为了性能考虑，如果要加日志字段，需要改变slice大小
		var fields = make([]elog.Field, 0, 15)
		var brokenPipe bool
		var event = "normal"
		defer func() {
			cost := time.Since(beg)

			fields = append(fields,
				elog.FieldType("http"), // GET, POST
				elog.FieldCost(cost),
				elog.FieldMethod(ctx.Req().Method+"."+ctx.Request.SelectedRoutePath()), // 完整路径
				elog.FieldAddr(ctx.Req().URL.RequestURI()),
				elog.FieldIP(ctx.ClientIP()),
				elog.FieldSize(int32(ctx.Response.ContentLength())),
				elog.FieldPeerIP(ctx.GetPeerIP()),
			)
			// 是否开启链路追踪，默认开启
			if config.EnableTraceInterceptor && opentracing.IsGlobalTracerRegistered() {
				fields = append(fields, elog.FieldTid(etrace.ExtractTraceID(ctx.Context.Context())))
			}

			config.mu.RLock()
			if config.EnableAccessInterceptorReq {
				fields = append(fields, elog.Any("req", map[string]interface{}{
					"metadata": ctx.Req().Header,
					"payload":  rb.String(),
				}))
			}

			if config.EnableAccessInterceptorRes && ok {
				fields = append(fields, elog.Any("res", map[string]interface{}{
					"metadata": ctx.Header(),
					"payload":  rw.body.String(),
				}))
			}
			config.mu.RUnlock()

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
					_ = ctx.WriteError(http.StatusInternalServerError, rec.(error)) // nolint: errcheck
				} else {
					ctx.WriteHeader(http.StatusInternalServerError)
				}

				event = "recover"
				stackInfo := stack(3)

				fields = append(fields,
					elog.FieldEvent(event),
					zap.ByteString("stack", stackInfo),
					elog.FieldErrAny(rec),
					elog.FieldCode(int32(ctx.StatusCode())),
					elog.FieldUniformCode(int32(ctx.StatusCode())),
				)
				logger.Error("access", fields...)
				return
			}
			if config.EnableAccessInterceptor {
				fields = append(fields,
					elog.FieldEvent(event),
					elog.FieldErrAny(ctx.Error()),
					elog.FieldCode(int32(ctx.StatusCode())),
				)
				logger.Info("access", fields...)
			}
		}()
		ctx.ProcessFilter()
	})
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
