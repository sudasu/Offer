package go_test

import "context"

type Demo struct {
	Data string
}

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -key=demoKey
	CacheDemos(c context.Context, keys []int64) (map[int64]*Demo, error)
	// redis: -key=demoKey -encode=json|gzip
	CacheDemo(c context.Context, key int64) (*Demo, error)
	// redis: -key=keyMid -encode=pb|gzip
	CacheDemo1(c context.Context, key int64, mid int64) (*Demo, error)
	// redis: -key=noneKey -encode=pb|gzip
	CacheNone(c context.Context) (*Demo, error)
	// redis: -key=demoKey
	CacheString(c context.Context, key int64) (string, error)

	// redis: -key=demoKey -expire=d.demoExpire -encode=json
	AddCacheDemos(c context.Context, values map[int64]*Demo) error
	// redis: -key=demo2Key -expire=d.demoExpire -encode=json
	AddCacheDemos2(c context.Context, values map[int64]*Demo, tp int64) error
	// 这里也支持自定义注释 会替换默认的注释
	// redis: -key=demoKey -expire=d.demoExpire -encode=json|gzip
	AddCacheDemo(c context.Context, key int64, value *Demo) error
	// redis: -key=keyMid -expire=d.demoExpire -encode=pb
	AddCacheDemo1(c context.Context, key int64, value *Demo, mid int64) error
	// redis: -key=noneKey -expire=d.demoExpire -encode=pb|gzip
	AddCacheNone(c context.Context, value *Demo) error
	// redis: -key=demoKey -expire=d.demoExpire
	AddCacheString(c context.Context, key int64, value string) error

	// redis: -key=demoKey
	DelCacheDemos(c context.Context, keys []int64) error
	// redis: -key=demoKey
	DelCacheDemo(c context.Context, key int64) error
	// redis: -key=keyMid
	DelCacheDemo1(c context.Context, key int64, mid int64) error
	// redis: -key=noneKey
	DelCacheNone(c context.Context) error
}
