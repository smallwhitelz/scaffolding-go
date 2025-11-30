package accesslog

import (
	"fmt"
	"net/http"
	"scaffolding-go/web"
	"testing"
)

func TestMiddlewareBuilder(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl := builder.LogFunc(func(log string) {
		fmt.Println(log)
	}).Build()
	server := web.NewHTTPServer(web.ServerWithMiddleware(mdl))
	server.Post("/a/b/*", func(ctx *web.Context) {
		fmt.Println("hello is me")
	})
	req, err := http.NewRequest(http.MethodPost, "/a/b/c", nil)
	if err != nil {
		t.Fatal(err)
	}
	server.ServeHTTP(nil, req)
}
