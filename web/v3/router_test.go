package v3

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter_AddRoute(t *testing.T) {
	// 第一个步骤是构造路由树
	// 第二个步骤是验证路由树
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail/:id",
		},
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}
	r := newRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	// 在这里断言路由树和你预期的一模一样
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
								paramChild: &node{
									path:    ":id",
									handler: mockHandler,
								},
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
						},
					},
				},
			},
			http.MethodPost: &node{
				path: "/",
				children: map[string]*node{
					"order": &node{
						path: "order",
						children: map[string]*node{
							"create": &node{
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": &node{
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}
	msg, ok := wantRouter.equal(&r)
	assert.True(t, ok, msg)
	// 这个是不行的，因为HandleFunc 是不可以比的
	//assert.Equal(t, wantRouter, r)

	r = newRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	}, "web: 路由必须以 / 开头")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	}, "web: 路由不能以 / 结尾")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b//c", mockHandler)
	}, "web: 路由中不能有连续的 / ")

	r = newRouter()
	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "web: 路由冲突, 路由已经存在")

	r = newRouter()
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	}, "web: 普通路由冲突, 路由已经存在")

	// 可用的 http method，要不要校验
	// 不用 因为我没必要暴露addRoute方法，用户只能调用我写的Get、Post等
	// mockHandler 为 nil？要不要校验
	// 不用 因为传了nil，相当于没注册，用户这样写没有意义
	//r.addRoute("aaa", "/a/b/c", nil)

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/*", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	}, "web:不允许同时注册路径参数和通配符路由，已有通配符路由")

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
	}, "web:不允许同时注册路径参数和通配符路由，已有路径参数")
}

// 返回一个错误信息，帮助我们排查问题
// bool 代表是否真的相等
func (r *router) equal(y *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的 http method"), false
		}
		msg, equal := v.equal(dst)
		if !equal {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return fmt.Sprintf("节点 path 不相等，want=%s, got=%s", n.path, y.path), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("节点 %s 的 children 长度不相等，want=%d, got=%d", n.path, len(n.children), len(y.children)), false
	}
	if n.starChild != nil {
		msg, ok := n.starChild.equal(y.starChild)
		if !ok {
			return msg, ok
		}
	}
	if n.paramChild != nil {
		msg, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return msg, ok
		}
	}
	// 比较handler
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if nHandler != yHandler {
		return fmt.Sprintf("节点 %s 的 handler 不相等", n.path), false
	}
	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("节点 %s 找不到对应的子节点 %s", n.path, path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "", true
}

func TestRouter_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodPost,
			path:   "/login/:username",
		},
	}
	r := newRouter()
	var mockHandler HandleFunc = func(ctx *Context) {}
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCases := []struct {
		name   string
		method string
		path   string

		wantFound bool
		info      *matchInfo
	}{
		{
			// 方法都不存在
			name:      "method not found",
			method:    http.MethodOptions,
			path:      "/order/detail",
			wantFound: false,
		},
		{
			// 完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			info: &matchInfo{
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
		{
			// 完全命中
			name:      "order star",
			method:    http.MethodGet,
			path:      "/order/abc",
			wantFound: true,
			info: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中了，但是没有handler
			name:      "order",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			info: &matchInfo{
				n: &node{
					path: "order",
					//handler: mockHandler,
					children: map[string]*node{
						"detail": &node{
							path:    "detail",
							handler: mockHandler,
						},
					},
				},
			},
		},
		{
			// 根节点
			name:      "root ",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: true,
			info: &matchInfo{
				n: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			// username路径参数匹配
			name:      "login username ",
			method:    http.MethodPost,
			path:      "/login/zl",
			wantFound: true,
			info: &matchInfo{
				n: &node{
					path:    ":username",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"username": "zl",
				},
			},
		},

		{
			// 没注册
			name:   "path not found ",
			method: http.MethodGet,
			path:   "/aaabbbccc",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.info.pathParams, info.pathParams)
			msg, ok := tc.info.n.equal(info.n)
			assert.True(t, ok, msg)
		})
	}
}
