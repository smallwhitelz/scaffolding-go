package v5

import (
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	// trees 是按照 HTTP 方法来组织的
	// 如 GET => *node
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 注册路由。
// 静态>正则>参数>通配符
// method 是 HTTP 方法
// - 已经注册了的路由，无法被覆盖。例如 /user/home 注册两次，会冲突
// - path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
// - 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 冲突
// - 不能在同一个位置同时注册通配符路由和参数路由，例如 /user/:id 和 /user/* 冲突
// - 同名路径参数，在路由匹配的时候，值会被覆盖。例如 /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
func (r *router) addRoute(method string, path string, handler HandleFunc) {
	// 先判断参数是否为空
	if path == "" {
		panic("web: 路由不能为空")
	}
	// 判断参数是否以 / 开始
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}
	// 判断参数是否以 / 结尾
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}
	// 先找到方法对应的树
	root, ok := r.trees[method]
	// 这是全新的一个http树
	if !ok {
		// 说明还没有创建这棵树
		// 创建根节点
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}
	// 根节点特殊处理
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突, 路由已经存在")
		}
		root.handler = handler
		return
	}
	// 处理path
	// 把 path 按照 / 切分
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		// 查找子节点，找不到就创建
		root = root.childOrCreate(seg)
	}
	// 最后把 handler 挂载上去
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handler
}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	// 先找到方法对应的树
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	// 根节点特殊处理
	if path == "/" {
		return &matchInfo{n: root}, true
	}
	// 处理path
	segs := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	for _, seg := range segs {
		var child *node
		child, ok = root.childOf(seg)
		if !ok {
			// 最后一段 *
			if root.typ == nodeTypeAny {
				mi.n = root
				return mi, true
			}
			return nil, false
		}
		if child.paramName != "" {
			mi.addValue(child.paramName, seg)
		}
		root = child
	}
	mi.n = root
	return mi, true
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是：
// 1. 静态完全匹配
// 2. 正则匹配，形式 :param_name(reg_expr)
// 3. 路径参数匹配：形式 :param_name
// 4. 通配符匹配：*
// 这是不回溯匹配
type node struct {
	typ nodeType

	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	paramChild *node
	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp
}

// child 返回子节点
// 第一个返回值 *node 是命中的节点
// 第二个返回值 bool 代表是否命中
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.childOfNonStatic(path)
	}
	child, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return child, true
}

// childOrCreate 查找子节点，
// 首先会判断 path 是不是通配符路径
// 其次判断 path 是不是参数路径，即以 : 开头的路径
// 最后会从 children 里面查找，
// 如果没有找到，那么会创建一个新的节点，并且保存在 node 里面
func (n *node) childOrCreate(path string) *node {
	// 判断是不是通配符路径
	if path == "*" {
		// 这里的判断是因为通配符我们认为优先级最低
		// 所以到这里要判断有没有其他优先级高的
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{
				path: path,
				typ:  nodeTypeAny,
			}
		}
		return n.starChild
	}
	// 以 : 开头，这里要判断是参数路由还是正则路由
	if path[0] == ':' {
		// 判断是正则还是参数
		paramName, expr, isReg := n.parseParamOrReg(path)
		if isReg {
			return n.childOrCreateReg(path, expr, paramName)
		}
		return n.childOrCreateParam(path, paramName)
	}
	// 静态路由
	if n.children == nil {
		n.children = map[string]*node{}
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{
			path: path,
			typ:  nodeTypeStatic,
		}
		n.children[path] = child
	}
	return child
}

func (n *node) childOrCreateReg(path string, expr string, paramName string) *node {
	if n.paramChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.regChild != nil {
		if n.regChild.regExpr.String() != expr || n.paramName != paramName {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("web: 正则表达式错误 %w", err))
		}
		n.regChild = &node{
			path:      path,
			paramName: paramName,
			regExpr:   regExpr,
			typ:       nodeTypeReg,
		}
	}
	return n.regChild
}

func (n *node) childOrCreateParam(path string, paramName string) *node {
	if n.regChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}
	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
		}
	} else {
		n.paramChild = &node{
			path:      path,
			paramName: paramName,
			typ:       nodeTypeParam,
		}
	}
	return n.paramChild
}

// parseParamOrReg 用于解析判断是不是正则表达式
// 第一个返回值是参数名字
// 第二个返回值是正则表达式
// 第三个返回值为 true 则说明是正则路由
func (n *node) parseParamOrReg(path string) (string, string, bool) {
	// 去除 :
	path = path[1:]
	// 判断是不是正则表达式
	// 正则表达式的格式是 :param_name(reg_expr)
	segs := strings.SplitN(path, "(", 2)
	if len(segs) == 2 {
		expr := segs[1]
		// 必须以 ) 结尾
		if strings.HasSuffix(expr, ")") {
			expr = expr[:len(expr)-1]
			return segs[0], expr, true
		}
	}
	return path, "", false
}

// childOfNonStatic 从非静态匹配的子节点里面查找
func (n *node) childOfNonStatic(path string) (*node, bool) {
	if n.regChild != nil {
		if n.regChild.regExpr.Match([]byte(path)) {
			return n.regChild, true
		}
	}
	if n.paramChild != nil {
		return n.paramChild, true
	}
	return n.starChild, n.starChild != nil
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		// 大多数情况，参数路径只会有一段
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}
