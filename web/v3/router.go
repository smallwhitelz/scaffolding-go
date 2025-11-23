package v3

import "strings"

// 用来支持对路由树的操作
// 代表路由树（森林）
// 很明显，不是并发安全的
// 但是没必要去处理，我们的要求是启动前用户已经注册好了路由
// 这里注册路由是单线程的
// 查找路由是多线程的
// 并发读不会有问题，并发读写会有问题
type router struct {
	// Beego Gin HTTP method 对应一棵树
	// GET 一棵树，POST也一棵树

	// http method => 路由树根节点
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 这里注册到路由树里面
// 用户写路由会有非常多的场景
// 所以我们要对path加一些限制
// path必须以 / 开头 ,中间不能有连续的 / ，不能以 / 结尾（除非只有 / 本身）
// 已经注册的路由，无法被覆盖，例如 /user 已经注册了，/user 不能再注册一次
// 不能在同一个位置同时注册路径参数和通配符路由
// 同名路径参数，在路由匹配的时候，值会被覆盖，例如 /order/:id/detail/:id，那么最后 ctx.PathParams["id"] 会是最后一个 id 的值
// 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 会冲突
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web: 路由不能为空")
	}
	// 首先找到树来
	root, ok := r.trees[method]
	if !ok {
		// 说明还没有创建这棵树
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}
	// 开头不能没有 /
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}
	// 结尾不能有 /
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	// 根节点特殊处理
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突, 路由已经存在")
		}
		root.handler = handleFunc
		return
	}
	path = path[1:] // 去掉开头的 /
	// 把 path 按照 / 切分
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		if seg == "" {
			panic("web: 路由中不能有连续的 / ")
		}
		// 递归下去，找准位置
		// 如果中途有节点不存在，你就要创建出来
		child := root.childOrCreate(seg)
		root = child
	}
	if root.handler != nil {
		panic("web: 普通路由冲突, 路由已经存在")
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	// 沿着树深度查找
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	// 根节点特殊处理
	if path == "/" {
		return &matchInfo{
			n: root,
		}, true
	}
	// 把前后 / 去掉
	path = strings.Trim(path, "/")
	// 切割
	segs := strings.Split(path, "/")
	var pathParams map[string]string
	for _, seg := range segs {
		child, paramChild, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		// 这是命中了路径参数
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			// path 是 :id 这种形式的
			pathParams[child.path[1:]] = seg
		}
		root = child
	}
	// 代表我确实有这个节点
	// 但是这个节点是不是用户注册的有 handler 的，就不一定了
	return &matchInfo{
		n:          root,
		pathParams: pathParams,
	}, true
}

func (n *node) childOrCreate(seg string) *node {
	if seg[0] == ':' {
		if n.starChild != nil {
			panic("web:不允许同时注册路径参数和通配符路由，已有通配符路由")
		}
		n.paramChild = &node{
			path: seg,
		}
		return n.paramChild
	}
	if seg == "*" {
		if n.paramChild != nil {
			panic("web:不允许同时注册路径参数和通配符路由，已有路径参数")
		}
		n.starChild = &node{
			path: seg,
		}
		return n.starChild
	}
	if n.children == nil {
		n.children = map[string]*node{}
	}
	res, ok := n.children[seg]
	if !ok {
		// 新建这个节点
		res = &node{path: seg}
		n.children[seg] = res
	}
	return res
}

// childOf 优先考虑静态匹配，匹配不上在考虑通配符匹配
// 第一个参数是子节点
// 第二个参数代表是否是路径参数匹配
// 第三个参数代表是否找到了节点
func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	child, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return child, false, ok
}

type node struct {
	path string
	// 子 path到子节点的映射
	children map[string]*node

	// 加一个通配符匹配
	starChild *node

	// 加一个路径参数
	paramChild *node

	// 缺一个代表用户注册的业务逻辑
	handler HandleFunc
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}
