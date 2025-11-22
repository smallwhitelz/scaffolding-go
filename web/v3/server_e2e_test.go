//go:build e2e

package v2

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := NewHTTPServer() // NewServer
	handler1 := func(ctx *Context) {
		fmt.Println("做第一件事")
	}
	handler2 := func(ctx *Context) {
		fmt.Println("做第二件事")
	}
	h.addRoute(http.MethodGet, "/user", func(ctx *Context) {
		handler1(ctx)
		handler2(ctx)
	})
	//h.Get("/user", func(ctx *Context) {
	//
	//})

	h.Get("/order/detail", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello world"))
	})

	h.Get("/order/abc", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello,%s", ctx.Req.URL.Path)))
	})
	// 方法一，完全委托给http包管理
	//http.ListenAndServe(":8081", h)
	//http.ListenAndServeTLS(":443", "", "", h)

	// 方法二 自己管
	h.Start(":8081")
}
