package article

import (
	"megin/pkg/context/api"
	"time"
)

// DetailOptions 定义文章详情接口的请求策略。
var DetailOptions = []api.RequestOption{
	// 文章详情：按文章 ID 从 Redis 读取缓存；未命中时执行接口并回填，缓存 1 分钟。
	api.WithCache("article", []string{"id"}, time.Minute),
}

// CreateOptions 定义文章创建接口的请求策略。
var CreateOptions = []api.RequestOption{
	// 创建文章：按当前登录用户和文章标题限速，5 秒内相同请求只允许一次；命中限速时返回“请求过于频繁”，不执行 Handler。
	api.WithRateLimit("article:create", []string{"title"}, true, 5*time.Second),
}

// UpdateOptions 定义文章更新接口的请求策略。
var UpdateOptions = []api.RequestOption{
	// 更新文章：按文章 ID 串行执行，避免并发更新覆盖，锁有效期为 5 秒；锁被占用时返回“当前数据正在处理中”，不执行 Handler。
	api.WithLock("article", []string{"id"}, 5*time.Second),
	// 更新成功后：删除该文章详情缓存，使后续详情请求回源最新数据。
	api.WithDeleteCache("article", []string{"id"}),
}

// DeleteOptions 定义文章删除接口的请求策略。
var DeleteOptions = []api.RequestOption{
	// 删除文章：按文章 ID 串行执行，避免并发删除同一条数据，锁有效期为 5 秒；锁被占用时返回“当前数据正在处理中”，不执行 Handler。
	api.WithLock("article", []string{"id"}, 5*time.Second),
	// 删除成功后：删除该文章详情缓存，避免继续读取已删除的数据。
	api.WithDeleteCache("article", []string{"id"}),
}
