//go:build e2e

package v1

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := &HTTPServer{} // NewServer
	handler1 := func(ctx Context) {
		fmt.Println("做第一件事")
	}
	handler2 := func(ctx Context) {
		fmt.Println("做第二件事")
	}
	h.AddRoute(http.MethodGet, "/user", func(ctx Context) {
		handler1(ctx)
		handler2(ctx)
	})
	h.Get("/user", func(ctx Context) {

	})
	// 方法一，完全委托给http包管理
	//http.ListenAndServe(":8081", h)
	//http.ListenAndServeTLS(":443", "", "", h)

	// 方法二 自己管
	h.Start(":8081")
}
