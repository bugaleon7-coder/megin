package bootstrap

import (
	"log"
	"megin/internal/cache"
	"megin/internal/config"
	"megin/internal/middleware"
	"megin/internal/migration"
	xrouter "megin/internal/router"
	"megin/internal/schedule"
	"megin/pkg/context/router"
	"megin/pkg/logger"
	"megin/pkg/validate"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 1,初始化服务
func ServerInit(configPath string, onStart func() error) {
	ServerInitWithMode(configPath, config.RunModeMixed, onStart)
}

// ServerInitWithMode 按指定运行模式初始化服务。
func ServerInitWithMode(configPath string, mode string, onStart func() error) {
	//1,解析配置文件
	conf := config.InitConfig(configPath, mode)
	//2,Log初始化
	logger.InitLog(logger.LogConfig{LogInConsole: true})
	//3,加载参数验证扩展
	validate.RegisterExtension()
	//4,仅首次安装时恢复全量初始化数据；已有数据库绝不执行初始化 SQL。
	if conf.Migrate.Enabled() {
		exists, err := migration.DatabaseExists(conf.Database.Dsn)
		if err != nil {
			log.Fatalln("检查初始化数据库失败:", err)
		}
		if !exists {
			if err := migration.ApplyInit(conf.Database.Dsn, conf.Migrate.OutputFile()); err != nil {
				log.Fatalln("执行初始化 SQL 失败:", err)
			}
		}
	}
	//5,数据库初始化
	config.InitDatabase(conf)
	//6,按日期顺序执行未执行的增量 SQL。
	if conf.Migrate.Enabled() {
		if _, err := migration.RunPending(config.GetMysqlDB(), conf.Database.Dsn, conf.Migrate.Directory(), conf.Migrate.OutputFile()); err != nil {
			log.Fatalln("执行增量 SQL 迁移失败:", err)
		}
	}
	//7,初始化默认缓存管理器，供限流、分布式锁等依赖 Redis 的能力统一复用。
	cache.InitManager(config.GetRedis().GetDB())
	//8,业务初始化
	err := onStart()
	if err != nil {
		log.Fatalln(err)
	}
}

func SetupTestRouter() *gin.Engine {
	//设为release,要不然输出的东西太多，影响视线
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.New()
	registry := router.NewRouteRegistry(ginRouter)
	registry.Use(middleware.Cors())
	registry.Use(middleware.TraceID())
	registry.Use(middleware.RequestLog())
	registry.Use(middleware.Recover())

	// 测试环境默认挂载前后台路由，便于统一验证接口行为。
	xrouter.InitGinModules(registry, xrouter.RouterModules{
		MountAPI:      true,
		MountAdminAPI: true,
	})
	return ginRouter
}

// 3,启动服务1
func ServerRun() {
	defer func() { _ = logger.Sync() }()
	conf := config.GetConfig()
	//1,Gin框架初始化
	route := xrouter.InitGinRouter(xrouter.RouterModules{
		MountAPI:      conf.GetRunMode() == config.RunModeMixed || conf.GetRunMode() == config.RunModeAPI,
		MountAdminAPI: conf.GetRunMode() == config.RunModeMixed || conf.GetRunMode() == config.RunModeAdminAPI,
	})
	schedule.Start()
	startPprofServer(conf)

	if route.Run(conf.ActiveListenAddr()) != nil {
		logger.Fatal("Server Run Error")
	}
}

// startPprofServer 在独立的本机地址启动运行时性能分析服务。
func startPprofServer(conf *config.ServiceConfig) {
	if !conf.Pprof.Enable {
		return
	}

	addr := conf.PprofListenAddr()
	go func() {
		logger.Info("Pprof Server Started", zap.String("addr", addr))
		logger.Info("Pprof 查看地址: http://" + addr + "/debug/pprof/")
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("Pprof Server Stopped", zap.Error(err))
		}
	}()
}
