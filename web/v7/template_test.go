package v7

import (
	"html/template"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginPage(t *testing.T) {

	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)
	engine := &GoTemplateEngine{
		T: tpl,
	}
	h := NewHTTPServer(ServerWithTemplateEngine(engine))
	h.Get("/login", func(ctx *Context) {
		err2 := ctx.Render("login.gohtml", nil)
		if err2 != nil {
			log.Println(err2)
		}
	})
	h.Start(":8081")
}
