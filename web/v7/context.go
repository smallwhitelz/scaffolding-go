package v7

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req *http.Request

	// Resp 如果用户直接使用这个
	// 那么他们就绕开了 RespData 和 RespStatusCode 两个
	// 那么部分 middleware 无法运作
	Resp http.ResponseWriter

	// 这两个主要是为了 middleware 读写用的
	RespData       []byte
	RespStatusCode int

	PathParams map[string]string

	queryValues url.Values

	MatchedRoute string

	tplEngine TemplateEngine
}

func (c *Context) Render(tplName string, data any) error {
	val, err := c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespStatusCode = http.StatusOK
	c.RespData = val
	return nil
}

func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespData = data
	c.RespStatusCode = status
	return err
}

// BindJSON 解决大部分人的需求即可
func (c *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("web: 输入为 nil")
	}
	if c.Req.Body == nil {
		return errors.New("web: 请求体为空")
	}
	decoder := json.NewDecoder(c.Req.Body)
	// 这两个方法很简单，用户如果需求，自己很简单的打开即可
	// 我们作为中间件的设计者没有必要去解决少数的需求
	// UserNumber => 数字就是用 json.Number 类型接收，避免精度丢失
	// 否则就是用 float64 接收
	//decoder.UseNumber()

	// 如果要是有一个未知字段，就会报错
	// 比如你 User 里面只有 Name 和 Age 两个字段
	// 但是你JSON额外多了一个 Gender 字段，那么就会报错
	//decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) (string, error) {
	// ParseForm 不会重复解析，因为是幂等的
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil
}

// 相比于表单，query没有缓存
func (c *Context) QueryValue(key string) (string, error) {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}
	vals, ok := c.queryValues[key]
	// 这样就可以区别出到底是空字符串，还是说确实没有这个key
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil
	// 无法区分是空字符串还是key不存在
	//return c.queryValues.Get(key), nil
}

func (c *Context) QueryValueV1(key string) StringValue {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}
	vals, ok := c.queryValues[key]
	// 这样就可以区别出到底是空字符串，还是说确实没有这个key
	if !ok {
		return StringValue{
			err: errors.New("web: key 不存在"),
		}
	}
	return StringValue{
		val: vals[0],
	}
	// 无法区分是空字符串还是key不存在
	//return c.queryValues.Get(key), nil
}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return val, nil
}

func (c *Context) PathValueV1(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{
			err: errors.New("web: key 不存在"),
		}
	}
	return StringValue{
		val: val,
	}
}

// 如果想提供对string类型的转换，可以这么玩
type StringValue struct {
	val string
	err error
}

// 达到链式调用的效果
func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}
