package v2

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

// 确保HTTPServer实现了Server接口
var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	Start(addr string) error
	//Start() error

	// addRoute 路由注册功能
	// method 是 HTTP 方法
	// path 是路由
	// handleFunc 是你的业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)
	// 这种允许注册多个，没有必要提供
	// 让用户自己去管
	// 要考虑如何中断，继续等问题；如果用户一个都不传，你怎么办
	// addRoute1(method string, path string, handles ...HandleFunc)
}

type HTTPSServer struct {
	HTTPServer
}

type HTTPServer struct {
	//router
	router
	//r *router
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Post(method string, path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Options(method string, path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodOptions, path, handleFunc)
}

// ServeHTTP 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑
	n, ok := h.router.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || n.handler == nil {
		// 路由没有命中，就是404
		ctx.Resp.WriteHeader(http.StatusNotFound)
		_, _ = ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}
	n.handler(ctx)
}

func (h *HTTPServer) Start(addr string) error {
	// 也可以自己创建server
	//http.Server{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 在这里可以让用户注册所谓的after start回调
	// 比如说往你的admin注册一下自己这个实例
	// 在这里执行一些你业务所需的前置条件
	return http.Serve(l, h)
}

func (h *HTTPServer) StartV1(addr string) error {
	return http.ListenAndServe(addr, h)
}
