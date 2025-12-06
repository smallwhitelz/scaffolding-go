package v6

import "sync"

// Middleware 函数式的责任链模式
// 函数式的洋葱模式
type Middleware func(next HandleFunc) HandleFunc

// MiddlewareV1 非函数式的中间件接口定义
type MiddlewareV1 interface {
	Invoke(next HandleFunc) HandleFunc
}

// Interceptor 变种玩法
// 这种Java中一般这么写
type Interceptor interface {
	Before(ctx *Context)
	After(ctx *Context)
	Surround(ctx *Context)
}

//type HandlerFuncV1 func(ctx *Context) (next bool)
//
//type ChainV1 struct {
//	handlers []HandlerFuncV1
//}
//
//func (c ChainV1) Run(ctx *Context) {
//	for _, h := range c.handlers {
//		next := h(ctx)
//		// 这种叫中断执行
//		if !next {
//			return
//		}
//	}
//}

type Net struct {
	handlers []HandlerFuncV1
}

func (c Net) Run(ctx *Context) {
	var wg sync.WaitGroup
	for _, hdl := range c.handlers {
		h := hdl
		if h.concurrent {
			wg.Add(1)
			go func() {
				h.Run(ctx)
				wg.Done()
			}()
		} else {
			h.Run(ctx)
		}
	}
	wg.Wait()
}

type HandlerFuncV1 struct {
	concurrent bool
	handlers   []*HandlerFuncV1
}

func (HandlerFuncV1) Run(ctx *Context) {

}
