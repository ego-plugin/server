package eref

import (
	"fmt"
	"github.com/gotomicro/ego/core/eflag"
	"github.com/gotomicro/ego/core/util/xtime"
	"sync"
	"time"
)

// Config HTTP config
type Config struct {
	Host                       string // IP地址，默认0.0.0.0
	Port                       int    // PORT端口，默认9001
	Network                    string
	ServerReadTimeout          time.Duration // 服务端，用于读取io报文过慢的timeout，通常用于互联网网络收包过慢，如果你的go在最外层，可以使用他，默认不启用。
	ServerReadHeaderTimeout    time.Duration // 服务端，用于读取io报文过慢的timeout，通常用于互联网网络收包过慢，如果你的go在最外层，可以使用他，默认不启用。
	ServerWriteTimeout         time.Duration // 服务端，用于读取io报文过慢的timeout，通常用于互联网网络收包过慢，如果你的go在最外层，可以使用他，默认不启用。
	ContextTimeout             time.Duration // 只能用于IO操作，才能触发，默认不启用
	EnableMetricInterceptor    bool          // 是否开启监控，默认开启
	EnableTraceInterceptor     bool          // 是否开启链路追踪，默认开启
	EnableLocalMainIP          bool          // 自动获取ip地址
	EnableGzip                 bool          //  开启gzip 压缩
	SlowLogThreshold           time.Duration // 服务慢日志，默认500ms
	EnableAccessInterceptor    bool          // 是否开启，记录请求数据
	EnableAccessInterceptorReq bool          // 是否开启记录请求参数，默认不开启
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数，默认不开启
	WebsocketHandshakeTimeout  time.Duration // 握手时间
	WebsocketReadBufferSize    int           // WebsocketReadBufferSize
	WebsocketWriteBufferSize   int           // WebsocketWriteBufferSize
	EnableWebsocketCompression bool          // 是否开通压缩
	EnableWebsocketCheckOrigin bool          // 是否支持跨域
	mu                         sync.RWMutex  // mutex for EnableAccessInterceptorReq、EnableAccessInterceptorRes、AccessInterceptorReqResFilter、aiReqResCelPrg
}

// DefaultConfig 反回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:                       eflag.String("host"),
		Port:                       9090,
		Network:                    "tcp",
		EnableAccessInterceptor:    true,
		EnableTraceInterceptor:     true,
		EnableMetricInterceptor:    true,
		SlowLogThreshold:           xtime.Duration("500ms"),
		EnableWebsocketCheckOrigin: false,
	}
}

// Address 反回地址
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
