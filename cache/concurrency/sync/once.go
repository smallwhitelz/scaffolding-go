package sync

import "sync"

type MyBiz struct {
	once sync.Once
}

func (m *MyBiz) Init() {
	m.once.Do(func() {
		// 初始化操作，只会执行一次
	})
}

type MyBusiness interface {
	DoSomething()
}

type singleton struct {
}

func (s *singleton) DoSomething() {
	panic("implement me")
}

var s *singleton
var singletonOnce sync.Once

// GetSingleton 懒加载
func GetSingleton() MyBusiness {
	singletonOnce.Do(func() {
		s = &singleton{}
	})
	return s
}

// 饥饿
func init() {
	// 用包初始化函数取代once
	s = &singleton{}
}
