package cache

import "fmt"

const (
	// CacheKeyFormatArticleCreateLimitByUser 创建文章接口按用户限流 key。
	// 请求字段：
	// - %d: 用户 ID
	CacheKeyFormatArticleCreateLimitByUser = "article:create:limit:user:%d"

	// CacheKeyFormatArticleCreateLimitByIP 创建文章接口按 IP 限流 key。
	// 请求字段：
	// - %s: 客户端 IP
	CacheKeyFormatArticleCreateLimitByIP = "article:create:limit:ip:%s"

	// CacheKeyFormatArticleUpdateLock 更新文章接口分布式锁 key。
	// 请求字段：
	// - %d: 文章 ID
	CacheKeyFormatArticleUpdateLock = "article:update:lock:%d"

	// CacheKeyFormatApiUserLoginToken C端用户登录态 token key。
	// 请求字段：
	// - %d: C端用户 ID
	CacheKeyFormatApiUserLoginToken = "api:user:login:token:%d"
)

// GetArticleCreateLimitKeyByUser 返回创建文章接口按用户限流的缓存 key。
func GetArticleCreateLimitKeyByUser(userID uint) string {
	return fmt.Sprintf(CacheKeyFormatArticleCreateLimitByUser, userID)
}

// GetArticleCreateLimitKeyByIP 返回创建文章接口按 IP 限流的缓存 key。
func GetArticleCreateLimitKeyByIP(clientIP string) string {
	return fmt.Sprintf(CacheKeyFormatArticleCreateLimitByIP, clientIP)
}

// GetArticleUpdateLockKey 返回更新文章接口的分布式锁 key。
func GetArticleUpdateLockKey(articleID int) string {
	return fmt.Sprintf(CacheKeyFormatArticleUpdateLock, articleID)
}

// GetApiUserLoginTokenKey 返回 C 端用户登录 token 的 Redis key。
func GetApiUserLoginTokenKey(userID uint) string {
	return fmt.Sprintf(CacheKeyFormatApiUserLoginToken, userID)
}
