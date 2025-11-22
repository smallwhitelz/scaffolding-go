package iris

import (
	"testing"

	"github.com/kataras/iris/v12"
)

func TestHelloWorld(t *testing.T) {
	app := iris.New()
	app.Get("/", func(ctx iris.Context) {
		_, _ = ctx.HTML("Hello <strong>%s</strong>!", "World")
	})
	app.Listen(":8083")
}
