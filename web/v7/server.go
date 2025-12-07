package v7

import (
	"fmt"
	"net/http"
)

type HandleFunc func(ctx *Context)

type Server interface {
	http.Handler
	// Start 启动服务器
	// addr 是监听地址。如果只指定端口，可以使用 ":8081"
	// 或者 "localhost:8082"
	Start(addr string) error

	// addRoute 注册一个路由
	// method 是 HTTP 方法
	addRoute(method string, path string, handler HandleFunc)
	// 我们并不采取这种设计方案
	// addRoute(method string, path string, handlers... HandleFunc)
}

// 确保 HTTPServer 肯定实现了 Server 接口
var _ Server = &HTTPServer{}

// HTTPServerOption Option模式
type HTTPServerOption func(*HTTPServer)

type HTTPServer struct {
	router
	mdls []Middleware

	log func(msg string, args ...any)

	tplEngine TemplateEngine
}

// NewHTTPServerV1 缺乏拓展性
//func NewHTTPServerV1(mdls ...Middleware) *HTTPServer {
//	return &HTTPServer{
//		router: newRouter(),
//		mdls:   mdls,
//	}
//}

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: newRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithTemplateEngine(tplEngine TemplateEngine) HTTPServerOption {
	return func(server *HTTPServer) {
		server.tplEngine = tplEngine
	}
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mdls = mdls
	}
}

// ServeHTTP HTTPServer 处理请求的入口
func (s *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:       request,
		Resp:      writer,
		tplEngine: s.tplEngine,
	}
	// 最后一个是这个
	root := s.serve
	// 然后这里就是利用最后一个不断往前回溯组装链条
	// 从后往前
	// 把后一个作为前一个next构造成链条
	for i := len(s.mdls) - 1; i >= 0; i-- {
		root = s.mdls[i](root)
	}
	// 这里执行的时候就是从前往后了
	// 这里最后一个步骤，就是考虑把 RespData 和 RespStatusCode 刷新到响应里
	// 第一个应该是回写响应的
	// 因为它在调用next之后才回写响应，
	// 所以实际上 flashResp 是最后一个步骤
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			// 就设置好了 RespData 和 RespStatusCode
			next(ctx)
			s.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (s *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || len(ctx.RespData) != n {
		s.log("写入响应失败 %v", err)
	}
}

// Start 启动服务器
func (s *HTTPServer) Start(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *HTTPServer) Post(path string, handler HandleFunc) {
	s.addRoute(http.MethodPost, path, handler)
}

func (s *HTTPServer) Get(path string, handler HandleFunc) {
	s.addRoute(http.MethodGet, path, handler)
}

func (s *HTTPServer) serve(ctx *Context) {
	mi, ok := s.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || mi.n == nil || mi.n.handler == nil {
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("NOT FOUND")
		return
	}
	ctx.PathParams = mi.pathParams
	ctx.MatchedRoute = mi.n.route
	mi.n.handler(ctx)
}
