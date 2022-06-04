package eref

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/util/xnet"
	"github.com/opentracing/opentracing-go"
)

// Container 容器
type Container struct {
	config *Config
	name   string
	logger *elog.Component
}

// DefaultContainer 默认容器
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load 加载配置key
func Load(key string) *Container {
	c := DefaultContainer()
	c.logger = c.logger.With(elog.FieldComponentName(key))
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	var (
		host string
		err  error
	)
	// 获取网卡ip
	if c.config.EnableLocalMainIP {
		host, _, err = xnet.GetLocalMainIP()
		if err != nil {
			host = ""
		}
		c.config.Host = host
	}
	c.name = key
	return c
}

// Build 构建组件
func (c *Container) Build() *Component {
	server := newComponent(c.name, c.config, c.logger)
	// 修正反代理IP
	restful.Filter(ProxyIpMiddleware(c.logger, c.config))
	// 错误恢复
	restful.Filter(recoverMiddleware(c.logger, c.config))
	if c.config.ContextTimeout > 0 {
		restful.Filter(timeoutMiddleware(c.config.ContextTimeout))
	}
	if c.config.EnableMetricInterceptor {
		restful.Filter(metricServerInterceptor())
	}

	if c.config.EnableTraceInterceptor && opentracing.IsGlobalTracerRegistered() {
		restful.Filter(traceServerInterceptor())
	}

	return server
}
