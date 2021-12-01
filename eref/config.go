package eref

import (
	"fmt"
	"github.com/gotomicro/ego/core/eflag"
	"github.com/gotomicro/ego/core/util/xtime"
	"time"
)

// Config HTTP config
type Config struct {
	Host                    string        // IP地址，默认0.0.0.0
	Port                    int           // PORT端口，默认9001
	EnableMetricInterceptor bool          // 是否开启监控，默认开启
	EnableTraceInterceptor  bool          // 是否开启链路追踪，默认开启
	EnableLocalMainIP       bool          // 自动获取ip地址
	EnableGzip              bool          //  开启gzip 压缩
	SlowLogThreshold        time.Duration // 服务慢日志，默认500ms

	EnableSwagger bool // 启用API文档

	WebsocketHandshakeTimeout  time.Duration // 握手时间
	WebsocketReadBufferSize    int
	WebsocketWriteBufferSize   int
	EnableWebsocketCompression bool // 是否开通压缩
	EnableWebsocketCheckOrigin bool // 是否支持跨域
}

// DefaultConfig 反回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:                       eflag.String("host"),
		Port:                       9090,
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
