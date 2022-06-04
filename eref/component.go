package eref

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/constant"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"net"
	"net/http"
	"strings"
	"sync"
)

// PackageName 包名
const PackageName = "server.eref"

// Component 构件
type Component struct {
	mu     sync.Mutex      // 互拆锁
	name   string          // 构件名称
	config *Config         // 配置
	logger *elog.Component // 日记

	Server           *http.Server      // HTTP 服务
	listener         net.Listener      // 网络地址
	routerCommentMap map[string]string // router 的中文注释，非并发安全
}

// newComponent 新建一个构件
func newComponent(name string, config *Config, logger *elog.Component) *Component {
	comp := &Component{
		name:             name,
		config:           config,
		logger:           logger,
		listener:         nil,
		routerCommentMap: make(map[string]string),
	}

	// 注册解析类型
	restful.RegisterEntityAccessor(MIME_MSGPACK, NewEntityAccessorMsgPack())
	restful.RegisterEntityAccessor(restful.MIME_JSON, NewEntityAccessorJson())
	return comp
}

// Name 配置名称
func (c *Component) Name() string {
	return c.name
}

// PackageName 包名
func (c *Component) PackageName() string {
	return PackageName
}

// Init 初始化
func (c *Component) Init() error {
	var err error
	c.listener, err = net.Listen("tcp", c.config.Address())
	if err != nil {
		c.logger.Panic("new eref server err", elog.FieldErrKind("listen err"), elog.FieldErr(err))
	}
	c.config.Port = c.listener.Addr().(*net.TCPAddr).Port
	return nil
}

// RegisterRouteComment 注册路由注释
func (c *Component) RegisterRouteComment(method, path, comment string) {
	c.routerCommentMap[commentUniqKey(method, path)] = comment
}

// Start implements server.Component interface.
func (c *Component) Start() error {
	for _, ws := range restful.DefaultContainer.RegisteredWebServices() {
		for _, route := range ws.Routes() {
			info, flag := c.routerCommentMap[commentUniqKey(route.Method, route.Path)]
			// 如果有注释，日志打出来
			if flag {
				c.logger.Info("add route", elog.FieldMethod(route.Method), elog.String("path", route.Path), elog.Any("info", info))
			} else {
				c.logger.Info("add route", elog.FieldMethod(route.Method), elog.String("path", route.Path))
			}
		}
	}
	// 因为start和stop在多个goroutine里，需要对Server上写锁
	c.mu.Lock()
	c.Server = &http.Server{
		Addr:              c.config.Address(),
		Handler:           restful.DefaultContainer,
		ReadHeaderTimeout: c.config.ServerReadHeaderTimeout,
		ReadTimeout:       c.config.ServerReadTimeout,
		WriteTimeout:      c.config.ServerWriteTimeout,
	}
	c.mu.Unlock()
	err := c.Server.Serve(c.listener)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Stop implements server.Component interface
// it will terminate go-restful server immediately
func (c *Component) Stop() error {
	c.mu.Lock()
	err := c.Server.Close()
	c.mu.Unlock()
	return err
}

// GracefulStop implements server.Component interface
// it will stop go-restful server gracefully
func (c *Component) GracefulStop(ctx context.Context) error {
	c.mu.Lock()
	err := c.Server.Shutdown(ctx)
	c.mu.Unlock()
	return err
}

// Info returns server info, used by governor and consumer balancer
func (c *Component) Info() *server.ServiceInfo {
	info := server.ApplyOptions(
		server.WithScheme("http"),
		server.WithAddress(c.listener.Addr().String()),
		server.WithKind(constant.ServiceProvider),
	)
	return &info
}

func commentUniqKey(method, path string) string {
	return fmt.Sprintf("%s@%s", strings.ToLower(method), path)
}

// Listener listener信息
func (c *Component) Listener() net.Listener {
	return c.listener
}
