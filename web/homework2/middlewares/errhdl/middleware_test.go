package errhdl

import (
	"net/http"
	"scaffolding-go/web"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := NewMiddleware()
	builder.AddCode(http.StatusNotFound, []byte(
		`<html>
	<h1>404 NOT FOUND</h1>
</html>`)).
		AddCode(http.StatusBadRequest, []byte(
			`<html>
	<h1>请求不对</h1>
</html>`))
	server := web.NewHTTPServer(web.ServerWithMiddleware(builder.Build()))
	server.Start(":8081")
}
