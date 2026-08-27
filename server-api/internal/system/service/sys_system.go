package service

import (
	"megin/internal/base"
	"megin/internal/config"
	systemDto "megin/internal/system/dto"
	"megin/pkg/context/api"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/goccy/go-yaml"
)

type SysSystem struct {
	base.Service
}

func NewSysSystem(ctx *api.Context) *SysSystem {
	s := &SysSystem{}
	s.Initialize(ctx)
	return s
}

func (s *SysSystem) GetSystemConfig() (*systemDto.SystemConfigResponse, error) {
	raw, err := loadConfigFileAsMap()
	if err != nil {
		return nil, s.Error(err, "读取系统配置失败")
	}

	view := defaultSystemConfigView()
	mergeAnyMap(view, raw)
	applyRuntimeConfigView(view)

	return &systemDto.SystemConfigResponse{Config: view}, nil
}

func (s *SysSystem) SetSystemConfig(req *systemDto.SystemConfigReq) error {
	raw, err := loadConfigFileAsMap()
	if err != nil {
		return s.Error(err, "读取系统配置失败")
	}

	mergeAnyMap(raw, req.Config)
	normalizeSystemConfig(raw)
	syncFrontConfigToLegacy(raw)
	applyInMemoryConfig(raw)

	if err := writeConfigFile(raw); err != nil {
		return s.Error(err, "写入系统配置失败")
	}
	return nil
}

func (s *SysSystem) GetServerInfo() (map[string]any, error) {
	return map[string]any{
		"os": map[string]any{
			"goos":         runtime.GOOS,
			"numCpu":       runtime.NumCPU(),
			"compiler":     runtime.Compiler,
			"goVersion":    runtime.Version(),
			"numGoroutine": runtime.NumGoroutine(),
		},
		"cpu": map[string]any{
			"cores": runtime.NumCPU(),
			"cpus":  make([]float64, runtime.NumCPU()),
		},
		"ram":  runtimeRAMInfo(),
		"disk": []map[string]any{runtimeDiskInfo()},
	}, nil
}

func (s *SysSystem) ReloadSystem() error {
	// 当前项目未接入热重载器，先保持接口可用并返回成功。
	return nil
}

func loadConfigFileAsMap() (map[string]any, error) {
	path := config.GetConfigPath(config.GetConfig().Env)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any)
	if err := yamlUnmarshal(content, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func writeConfigFile(data map[string]any) error {
	path := config.GetConfigPath(config.GetConfig().Env)
	content, err := yamlMarshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func defaultSystemConfigView() map[string]any {
	conf := config.GetConfig()
	port := 0
	if v, err := strconv.Atoi(conf.Port); err == nil {
		port = v
	}

	return map[string]any{
		"app": map[string]any{
			"name":     conf.App.Name,
			"env":      conf.App.Env,
			"debug":    conf.App.Debug,
			"version":  conf.App.Version,
			"file_url": conf.App.FileUrl,
			"mode":     conf.GetRunMode(),
		},
		"servers": map[string]any{
			"mixed": map[string]any{
				"port": conf.Servers.Mixed.Port,
			},
			"api": map[string]any{
				"port": conf.Servers.API.Port,
			},
			"admin_api": map[string]any{
				"port": conf.Servers.AdminAPI.Port,
			},
		},
		"service_name": conf.ServiceName,
		"ip":           conf.IP,
		"port":         conf.Port,
		"debug":        conf.Debug,
		"env":          conf.Env,
		"version":      conf.Version,
		"file_url":     conf.FileUrl,
		"database": map[string]any{
			"driver": conf.Database.Driver,
			"dsn":    conf.Database.Dsn,
		},
		"jwt": map[string]any{
			"secret":         conf.Jwt.Secret,
			"expire_seconds": conf.Jwt.ExpireSeconds,
		},
		"redis": map[string]any{
			"addr":     conf.Redis.Addr,
			"password": conf.Redis.Password,
		},
		"totp": map[string]any{
			"enable": conf.TOTP.Enable,
			"issuer": conf.TOTP.Issuer,
		},
		"system": map[string]any{
			"addr":                 port,
			"db-type":              conf.Database.Driver,
			"oss-type":             "local",
			"use-multipoint":       false,
			"use-redis":            conf.Redis.Addr != "",
			"use-mongo":            false,
			"use-strict-auth":      conf.System.UseStrictAuth,
			"iplimit-count":        0,
			"iplimit-time":         0,
			"disable-auto-migrate": false,
			"router-prefix":        "",
		},
		"zap": map[string]any{
			"level":          "info",
			"format":         "console",
			"prefix":         "",
			"director":       "logs",
			"encode-level":   "LowercaseColorLevelEncoder",
			"stacktrace-key": "stacktrace",
			"retention-day":  7,
			"show-line":      false,
			"log-in-console": true,
		},
		"mysql":  map[string]any{},
		"mssql":  map[string]any{},
		"sqlite": map[string]any{},
		"pgsql":  map[string]any{},
		"oracle": map[string]any{},
		"excel":  map[string]any{},
		"autocode": map[string]any{
			"transfer-restart":  false,
			"root":              "",
			"server":            "",
			"server-api":        "",
			"server-initialize": "",
			"server-model":      "",
			"server-request":    "",
			"server-router":     "",
			"server-service":    "",
			"web":               "",
			"web-api":           "",
			"web-form":          "",
			"web-table":         "",
		},
		"mongo": map[string]any{
			"coll":               "",
			"options":            "",
			"database":           "",
			"username":           "",
			"password":           "",
			"min-pool-size":      0,
			"max-pool-size":      0,
			"socket-timeout-ms":  0,
			"connect-timeout-ms": 0,
			"is-zap":             false,
			"hosts": []map[string]any{
				{"host": "", "port": ""},
			},
		},
		"qiniu":         map[string]any{},
		"tencent-cos":   map[string]any{},
		"aliyun-oss":    map[string]any{},
		"hua-wei-obs":   map[string]any{},
		"cloudflare-r2": map[string]any{},
		"minio":         map[string]any{},
		"captcha":       map[string]any{},
		"local":         map[string]any{},
		"email":         map[string]any{},
		"timer": map[string]any{
			"detail": map[string]any{},
		},
		"api-doc": map[string]any{
			"enable":                     true,
			"output-swagger-file":        "static/swagger/api/swagger.json",
			"api-output-swagger-file":    "static/swagger/api/swagger.json",
			"admin-output-swagger-file":  "static/swagger/admin-api/swagger.json",
			"system-output-swagger-file": "static/swagger/system/swagger.json",
			"gen-scan-dir": []string{
				"pkg/context/api",
				"internal/api",
				"internal/admin-api",
				"internal/system",
			},
		},
	}
}

func applyRuntimeConfigView(view map[string]any) {
	if jwt, ok := view["jwt"].(map[string]any); ok {
		if secret, ok := jwt["secret"]; ok {
			jwt["signing-key"] = secret
		}
		if expires, ok := jwt["expire_seconds"]; ok {
			jwt["expires-time"] = expires
		}
		jwt["buffer-time"] = 0
		jwt["issuer"] = "xadmin"
	}
	if system, ok := view["system"].(map[string]any); ok {
		system["use-strict-auth"] = config.GetConfig().System.UseStrictAuth
	}
}

func normalizeSystemConfig(raw map[string]any) {
	// 保证核心分支存在，避免写入后被前端二次读取时出现空对象访问错误。
	ensureMap(raw, "system")
	ensureMap(raw, "jwt")
	ensureMap(raw, "redis")
	ensureMap(raw, "database")
	ensureMap(raw, "totp")
	ensureMap(raw, "api-doc")
	ensureMap(raw, "app")
	ensureMap(raw, "servers")
}

func syncFrontConfigToLegacy(raw map[string]any) {
	if app, ok := getNestedMap(raw, "app"); ok {
		if name, ok := app["name"].(string); ok && name != "" {
			if mixed, ok := getNestedMap(raw, "servers", "mixed"); ok {
				if _, exists := mixed["service_name"]; !exists {
					mixed["service_name"] = name + "-" + config.RunModeMixed
				}
			}
		}
		if env, ok := app["env"].(string); ok && env != "" {
			raw["env"] = env
		}
		if debug, ok := app["debug"].(bool); ok {
			raw["debug"] = debug
		}
		if version, ok := app["version"].(string); ok && version != "" {
			raw["version"] = version
		}
		if fileURL, ok := app["file_url"].(string); ok {
			raw["file_url"] = fileURL
		}
	}

	activeMode := config.GetConfig().GetRunMode()
	if servers, ok := getNestedMap(raw, "servers", activeMode); ok {
		if serviceName, ok := servers["service_name"].(string); ok && serviceName != "" {
			raw["service_name"] = serviceName
		}
		if ip, ok := servers["ip"].(string); ok && ip != "" {
			raw["ip"] = ip
		}
		if port, ok := servers["port"].(string); ok && port != "" {
			raw["port"] = port
		}
	}

	if system, ok := getNestedMap(raw, "system"); ok {
		if addr := asInt(system["addr"]); addr > 0 {
			raw["port"] = strconv.Itoa(addr)
			if servers, ok := getNestedMap(raw, "servers", activeMode); ok {
				servers["port"] = raw["port"]
			}
		}
		if dbType, ok := system["db-type"].(string); ok && dbType != "" {
			if database, ok := getNestedMap(raw, "database"); ok {
				database["driver"] = dbType
			}
		}
	}

	if jwt, ok := getNestedMap(raw, "jwt"); ok {
		if signingKey, ok := jwt["signing-key"].(string); ok && signingKey != "" {
			jwt["secret"] = signingKey
		}
		if expires := asInt64(jwt["expires-time"]); expires > 0 {
			jwt["expire_seconds"] = expires
		}
	}

	if redis, ok := getNestedMap(raw, "redis"); ok {
		if addr, ok := redis["addr"].(string); ok && addr != "" {
			redis["addr"] = addr
		}
		if password, ok := redis["password"].(string); ok {
			redis["password"] = password
		}
	}
}

func applyInMemoryConfig(raw map[string]any) {
	conf := config.GetConfig()

	if v, ok := getNestedMap(raw, "app"); ok {
		if name, ok := v["name"].(string); ok && name != "" {
			conf.App.Name = name
		}
		if env, ok := v["env"].(string); ok && env != "" {
			conf.App.Env = env
			conf.Env = env
		}
		if debug, ok := v["debug"].(bool); ok {
			conf.App.Debug = debug
			conf.Debug = debug
		}
		if version, ok := v["version"].(string); ok && version != "" {
			conf.App.Version = version
			conf.Version = version
		}
		if fileURL, ok := v["file_url"].(string); ok {
			conf.App.FileUrl = fileURL
			conf.FileUrl = fileURL
		}
		if mode, ok := v["mode"].(string); ok && mode != "" {
			conf.App.Mode = mode
		}
	}

	applyServerNode := func(target *config.ServerNode, values map[string]any) {
		if port, ok := values["port"].(string); ok && port != "" {
			target.Port = port
		}
	}
	if v, ok := getNestedMap(raw, "servers", config.RunModeMixed); ok {
		applyServerNode(&conf.Servers.Mixed, v)
	}
	if v, ok := getNestedMap(raw, "servers", config.RunModeAPI); ok {
		applyServerNode(&conf.Servers.API, v)
	}
	if v, ok := getNestedMap(raw, "servers", config.RunModeAdminAPI); ok {
		applyServerNode(&conf.Servers.AdminAPI, v)
	}

	if v, ok := getNestedMap(raw, "system"); ok {
		if addr := asInt(v["addr"]); addr > 0 {
			conf.Port = strconv.Itoa(addr)
			conf.ActiveServer().Port = conf.Port
		}
		if dbType, ok := v["db-type"].(string); ok && dbType != "" {
			conf.Database.Driver = dbType
		}
		if strictAuth, ok := v["use-strict-auth"].(bool); ok {
			conf.System.UseStrictAuth = strictAuth
		}
	}

	if v, ok := getNestedMap(raw, "jwt"); ok {
		if secret, ok := v["signing-key"].(string); ok && secret != "" {
			conf.Jwt.Secret = secret
		} else if secret, ok := v["secret"].(string); ok && secret != "" {
			conf.Jwt.Secret = secret
		}
		if expires := asInt64(v["expires-time"]); expires > 0 {
			conf.Jwt.ExpireSeconds = expires
		} else if expires := asInt64(v["expire_seconds"]); expires > 0 {
			conf.Jwt.ExpireSeconds = expires
		}
	}

	if v, ok := getNestedMap(raw, "redis"); ok {
		if addr, ok := v["addr"].(string); ok && addr != "" {
			conf.Redis.Addr = addr
		}
		if password, ok := v["password"].(string); ok {
			conf.Redis.Password = password
		}
	}

	if v, ok := getNestedMap(raw, "totp"); ok {
		if enable, ok := v["enable"].(bool); ok {
			conf.TOTP.Enable = enable
		}
		if issuer, ok := v["issuer"].(string); ok {
			conf.TOTP.Issuer = issuer
		}
	}

	active := conf.ActiveServer()
	conf.ServiceName = conf.ActiveServiceName()
	conf.IP = "0.0.0.0"
	conf.Port = active.Port
}

func runtimeRAMInfo() map[string]any {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	usedMb := int64(mem.Alloc / 1024 / 1024)
	totalMb := int64(mem.Sys / 1024 / 1024)
	usedPercent := 0.0
	if totalMb > 0 {
		usedPercent = float64(usedMb) * 100 / float64(totalMb)
	}
	return map[string]any{
		"totalMb":     totalMb,
		"usedMb":      usedMb,
		"usedPercent": usedPercent,
	}
}

func runtimeDiskInfo() map[string]any {
	path := "."
	if wd, err := os.Getwd(); err == nil {
		path = wd
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return map[string]any{
			"mountPoint":  path,
			"totalMb":     0,
			"usedMb":      0,
			"totalGb":     0,
			"usedGb":      0,
			"usedPercent": 0,
		}
	}
	total := int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - free
	usedPercent := 0.0
	if total > 0 {
		usedPercent = float64(used) * 100 / float64(total)
	}
	return map[string]any{
		"mountPoint":  path,
		"totalMb":     total / 1024 / 1024,
		"usedMb":      used / 1024 / 1024,
		"totalGb":     total / 1024 / 1024 / 1024,
		"usedGb":      used / 1024 / 1024 / 1024,
		"usedPercent": usedPercent,
	}
}

func mergeAnyMap(dst map[string]any, src map[string]any) {
	for k, v := range src {
		if v == nil {
			continue
		}
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				mergeAnyMap(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

func getNestedMap(m map[string]any, keys ...string) (map[string]any, bool) {
	current := m
	for _, key := range keys {
		v, ok := current[key]
		if !ok {
			return nil, false
		}
		mv, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		current = mv
	}
	return current, true
}

func ensureMap(m map[string]any, key string) map[string]any {
	if mv, ok := getNestedMap(m, key); ok {
		return mv
	}
	mv := map[string]any{}
	m[key] = mv
	return mv
}

func asInt(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int8:
		return int(t)
	case int16:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint:
		return int(t)
	case uint8:
		return int(t)
	case uint16:
		return int(t)
	case uint32:
		return int(t)
	case uint64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(t))
		return i
	default:
		return 0
	}
}

func asInt64(v any) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int8:
		return int64(t)
	case int16:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	case uint:
		return int64(t)
	case uint8:
		return int64(t)
	case uint16:
		return int64(t)
	case uint32:
		return int64(t)
	case uint64:
		return int64(t)
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		i, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return i
	default:
		return 0
	}
}

func yamlUnmarshal(content []byte, out any) error {
	return yaml.Unmarshal(content, out)
}

func yamlMarshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}
