package beego

import (
	"testing"

	"github.com/beego/beego/v2/server/web"
)

func TestUserController(t *testing.T) {
	web.BConfig.CopyRequestBody = true
	c := &UserController{}
	web.Router("/user", c, "get:GetUser")
	web.Run(":8081")
}
