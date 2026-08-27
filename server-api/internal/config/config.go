package config

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/goccy/go-yaml"
)

const (
	ConfigDir        = "./config/"
	ConfigFileFormat = "config-%s.yaml"
)

const (
	EnvProd = "prod"
	EnvTest = "test"
	EnvDev  = "dev"
)

const (
	RunModeMixed    = "mixed"
	RunModeAPI      = "api"
	RunModeAdminAPI = "admin_api"
)

type AppConfig struct {
	Name    string `yaml:"name"`
	Env     string `yaml:"env"`
	Debug   bool   `yaml:"debug"`
	Version string `yaml:"version"`
	FileUrl string `yaml:"file_url"`
	Mode    string `yaml:"mode"`
}

type Database struct {
	Dsn    string `yaml:"dsn"`
	Driver string `yaml:"driver"`
}

type Redis struct {
	Addr     string `yaml:"addr"` //127.0.0.1:6379
	Password string `yaml:"password"`
}

type JwtConfig struct {
	Secret        string `yaml:"secret"`
	ExpireSeconds int64  `yaml:"expire_seconds"`
}

type TOTPConfig struct {
	Enable bool   `yaml:"enable"`
	Issuer string `yaml:"issuer"`
}

// CaptchaConfig 是后台登录图形验证码配置。
type CaptchaConfig struct {
	Enable bool `yaml:"enable"` // 是否启用后台登录图形验证码；关闭后验证码接口不生成图片，登录接口也不校验验证码
}

// AdminConfig 是仅作用于管理后台的安全与界面配置。
type AdminConfig struct {
	TOTP          TOTPConfig    `yaml:"totp"`
	Captcha       CaptchaConfig `yaml:"captcha"`
	UseStrictAuth bool          `yaml:"use-strict-auth"`
	Watermark     bool          `yaml:"watermark"`
}

// RateLimitRule 是单个维度的令牌桶配置。
type RateLimitRule struct {
	Rate  float64 `yaml:"rate"`  // 每秒生成的令牌数，支持小数，例如 0.5 表示每两秒生成一个令牌
	Burst int     `yaml:"burst"` // 令牌桶容量，决定空闲后允许瞬时通过的最大请求数
}

// UIDRateLimitConfig 是登录用户维度的令牌桶配置。
type UIDRateLimitConfig struct {
	Enable        bool             `yaml:"enable"` // 首次建表时是否启用全局 UID 初始规则；建表后以 MySQL 规则状态为准
	RateLimitRule `yaml:",inline"` // 首次建表时写入 MySQL 的全局 UID 初始规则参数
}

// APIRateLimitConfig 是前台 API 的内存限流配置。
type APIRateLimitConfig struct {
	Enable                 bool               `yaml:"enable"`                   // 是否启用 /api 路由的进程内限流
	IP                     RateLimitRule      `yaml:"ip"`                       // 首次建表时写入 MySQL 的全局 IP 初始规则参数
	UID                    UIDRateLimitConfig `yaml:"uid"`                      // 首次建表时写入 MySQL 的全局 UID 初始规则配置
	IdleExpirationSeconds  int64              `yaml:"idle-expiration-seconds"`  // IP 或 UID 令牌桶持续空闲多久后允许从内存删除
	CleanupIntervalSeconds int64              `yaml:"cleanup-interval-seconds"` // 两次惰性清理之间的最短间隔，清理由后续请求触发
	RefreshIntervalSeconds int64              `yaml:"refresh-interval-seconds"` // 从 MySQL 定时刷新规则的周期秒数
}

type ApiDoc struct {
	Enable                  bool     `yaml:"enable"`                     //是否开启文档功能
	GenScanDir              []string `yaml:"gen-scan-dir"`               // 生成文档时扫描目录,常见如dto,model,handler等声明接口函数和对象的地方
	OutputSwaggerFile       string   `yaml:"output-swagger-file"`        // 兼容旧配置
	ApiOutputSwaggerFile    string   `yaml:"api-output-swagger-file"`    // 前端 API 文档输出文件
	AdminOutputSwaggerFile  string   `yaml:"admin-output-swagger-file"`  // 后台 Admin API 文档输出文件
	SystemOutputSwaggerFile string   `yaml:"system-output-swagger-file"` // 系统管理 API 文档输出文件
}

// PprofConfig 是运行时性能分析服务配置。
type PprofConfig struct {
	Enable bool   `yaml:"enable"`
	Port   string `yaml:"port"`
}

type ServerNode struct {
	Port string `yaml:"port"`
}

type ServersConfig struct {
	Mixed    ServerNode `yaml:"mixed"`
	API      ServerNode `yaml:"api"`
	AdminAPI ServerNode `yaml:"admin_api"`
}

// 服务端配置
type ServiceConfig struct {
	App          AppConfig          `yaml:"app"`
	Servers      ServersConfig      `yaml:"servers"`
	ServiceName  string             `yaml:"service_name"`
	IP           string             `yaml:"ip"`
	Port         string             `yaml:"port"`
	Debug        bool               `yaml:"debug"`
	Env          string             `yaml:"env"`
	Version      string             `yaml:"version"`
	FileUrl      string             `yaml:"file_url"`
	Database     Database           `yaml:"database"`
	Jwt          JwtConfig          `yaml:"jwt"`
	Redis        Redis              `yaml:"redis"`
	Admin        AdminConfig        `yaml:"admin"`
	APIRateLimit APIRateLimitConfig `yaml:"api-rate-limit"`
	ApiDoc       ApiDoc             `yaml:"api-doc"`
	Pprof        PprofConfig        `yaml:"pprof"`
}

func (config *ServiceConfig) GetRunMode() string {
	mode := config.App.Mode
	if mode == "" {
		mode = RunModeMixed
	}
	switch mode {
	case RunModeMixed, RunModeAPI, RunModeAdminAPI:
		return mode
	default:
		return RunModeMixed
	}
}

func (config *ServiceConfig) ActiveServer() *ServerNode {
	switch config.GetRunMode() {
	case RunModeAPI:
		return &config.Servers.API
	case RunModeAdminAPI:
		return &config.Servers.AdminAPI
	default:
		return &config.Servers.Mixed
	}
}

func (config *ServiceConfig) ActiveListenAddr() string {
	server := config.ActiveServer()
	port := server.Port
	if port == "" {
		port = defaultPortByMode(config.GetRunMode())
	}
	return "0.0.0.0:" + port
}

// PprofListenAddr 返回性能分析服务监听地址。
func (config *ServiceConfig) PprofListenAddr() string {
	port := config.Pprof.Port
	if port == "" {
		port = "6060"
	}
	return "127.0.0.1:" + port
}

func (config *ServiceConfig) ActiveServiceName() string {
	if config.App.Name == "" {
		return "shop-api-" + config.GetRunMode()
	}
	return config.App.Name + "-" + config.GetRunMode()
}

func (config *ServiceConfig) IsProdEnv() bool {
	if config.Env != EnvProd {
		return false
	}
	return true
}

func (config *ServiceConfig) IsDevEnv() bool {
	if config.Env != EnvDev {
		return false
	}
	return true
}

func (config *ServiceConfig) IsDebug() bool {
	if config.IsProdEnv() {
		return false
	}
	if config.Debug == true {
		return true
	}
	return false
}

var config = new(ServiceConfig)

func GetConfig() *ServiceConfig {
	return config
}

func GetConfigPath(env string) string {
	return path.Join(ConfigDir, fmt.Sprintf(ConfigFileFormat, env))
}

func InitConfig(path string, mode string) *ServiceConfig {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("InitConfig: ", path, ".  config err: ", err.Error())
		return nil
	}

	if err = yaml.Unmarshal(content, config); err != nil {
		log.Fatal("InitConfig: ", path, ".  config err: ", err.Error())
		return nil
	}
	config.normalize(mode)
	return config
}

func (config *ServiceConfig) normalize(mode string) {
	if config.App.Env == "" {
		config.App.Env = config.Env
	}
	if config.App.Version == "" {
		config.App.Version = config.Version
	}
	if config.App.FileUrl == "" {
		config.App.FileUrl = config.FileUrl
	}
	if config.App.Name == "" {
		config.App.Name = config.ServiceName
	}
	if config.App.Mode == "" {
		config.App.Mode = RunModeMixed
	}
	if mode != "" {
		config.App.Mode = mode
	}

	config.applyServerDefaults(&config.Servers.Mixed, RunModeMixed)
	config.applyServerDefaults(&config.Servers.API, RunModeAPI)
	config.applyServerDefaults(&config.Servers.AdminAPI, RunModeAdminAPI)
	config.applyAPIRateLimitDefaults()

	active := config.ActiveServer()
	config.Env = config.App.Env
	config.Debug = config.App.Debug
	config.Version = config.App.Version
	config.FileUrl = config.App.FileUrl
	config.ServiceName = config.ActiveServiceName()
	config.IP = "0.0.0.0"
	config.Port = active.Port
	if config.Port == "" {
		config.Port = defaultPortByMode(config.GetRunMode())
	}
}

func (config *ServiceConfig) applyAPIRateLimitDefaults() {
	// 配置缺失或值不合法时使用保守默认值，避免创建无法正常工作的令牌桶。
	config.APIRateLimit.Enable = true
	config.APIRateLimit.UID.Enable = true
	if config.APIRateLimit.IP.Rate <= 0 {
		config.APIRateLimit.IP.Rate = 20
	}
	if config.APIRateLimit.IP.Burst <= 0 {
		config.APIRateLimit.IP.Burst = 40
	}
	if config.APIRateLimit.UID.Rate <= 0 {
		config.APIRateLimit.UID.Rate = 10
	}
	if config.APIRateLimit.UID.Burst <= 0 {
		config.APIRateLimit.UID.Burst = 20
	}
	if config.APIRateLimit.IdleExpirationSeconds <= 0 {
		config.APIRateLimit.IdleExpirationSeconds = 600
	}
	if config.APIRateLimit.CleanupIntervalSeconds <= 0 {
		config.APIRateLimit.CleanupIntervalSeconds = 300
	}
	if config.APIRateLimit.RefreshIntervalSeconds <= 0 {
		config.APIRateLimit.RefreshIntervalSeconds = 60
	}
}

func (config *ServiceConfig) applyServerDefaults(node *ServerNode, mode string) {
	if node.Port == "" {
		node.Port = defaultPortByMode(mode)
	}
}

func defaultPortByMode(mode string) string {
	switch mode {
	case RunModeAPI:
		return "8801"
	case RunModeAdminAPI:
		return "8802"
	default:
		return "8800"
	}
}
