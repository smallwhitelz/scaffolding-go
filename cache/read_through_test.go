package cache

type User struct {
	Name string
}

//func TestReadThroughCacheV1_Get(t *testing.T) {
//	// 这种要多一层类型断言，但是可以断言多个类型
//	var c1 ReadThroughCache = ReadThroughCache{
//		LoadFunc: func(ctx context.Context, key string) (any, error) {
//			if strings.HasPrefix(key, "user") {
//				// 加载user
//				return nil, nil
//			} else if strings.HasPrefix(key, "order") {
//				// 加载order
//				return nil, nil
//			} else {
//				return nil, errors.New("不知道什么类型")
//			}
//		},
//	}
//	// 这种虽然不用类型断言，但是一个声明只能用于一个类型
//	var c2 ReadThroughCacheV1[User]
//
//	res, _ := c1.Get(context.Background(), "user")
//	u := res.(User)
//	//res, _ = c1.Get(context.Background(), "order")
//	//o:=res.(Order)
//
//	u, _ = c2.Get(context.Background(), "user")
//	t.Log(u.Name)
//}
