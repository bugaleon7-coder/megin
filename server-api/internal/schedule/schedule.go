package schedule

import (
	"megin/internal/config"
	rateLimitRuntime "megin/internal/system/runtime"
	"megin/pkg/logger"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func goroutineWatcher() {
	num := runtime.NumGoroutine()
	logger.Info("NumGoroutine", zap.Int("num", num))
}

func Start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("schedule defer Error", zap.Any("err", err))
			debug.PrintStack()
		}
	}()

	schedule := cron.New(cron.WithSeconds())
	schedule.AddFunc("@every 1m", goroutineWatcher)
	refreshInterval := time.Duration(config.GetConfig().APIRateLimit.RefreshIntervalSeconds) * time.Second
	if refreshInterval > 0 {
		schedule.AddFunc("@every "+refreshInterval.String(), func() {
			if _, err := rateLimitRuntime.ReloadDefault(); err != nil {
				logger.Error("刷新API限流规则失败", zap.Error(err))
			}
		})
	}
	schedule.Start()
}
