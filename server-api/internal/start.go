package internal

import (
	"megin/internal/config"
	rateLimitRuntime "megin/internal/system/runtime"
	"megin/pkg/logger"
	"time"
)

// 服务启动后，初始化相关操作可以放在这里执行,比如全局变量，订阅kafka
func OnServerStart() error {
	logger.Info("OnServerStart Run....")
	conf := config.GetConfig()
	db := config.GetMysqlDB()
	_, err := rateLimitRuntime.InitDefaultManager(
		db,
		time.Duration(conf.APIRateLimit.IdleExpirationSeconds)*time.Second,
		time.Duration(conf.APIRateLimit.CleanupIntervalSeconds)*time.Second,
	)
	return err
}
