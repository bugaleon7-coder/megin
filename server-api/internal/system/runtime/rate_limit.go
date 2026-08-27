package runtime

import (
	"fmt"
	rateLimitModel "megin/internal/system/model"
	rateLimitRepo "megin/internal/system/repository"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type RouteKey struct {
	Method string
	Path   string
}

type compiledRule struct {
	ID          uint
	Limit       rate.Limit
	Burst       int
	Fingerprint string
}

type routeRules struct {
	IP  *compiledRule
	UID *compiledRule
}

type snapshot struct {
	GlobalIP  *compiledRule
	GlobalUID *compiledRule
	Routes    map[RouteKey]routeRules
	RuleCount int
}

type bucketKey struct {
	RuleID  uint
	Subject string
}

type bucketEntry struct {
	limiter     *rate.Limiter
	fingerprint string
	lastSeen    time.Time
}

// Decision 是一次内存限流判断结果。
type Decision struct {
	Allowed    bool
	RetryAfter time.Duration
	RuleID     uint
}

// Manager 管理只读规则快照和每个规则对应的实时内存令牌桶。
type Manager struct {
	db              *gorm.DB
	snapshot        atomic.Pointer[snapshot]
	reloadMu        sync.Mutex
	bucketMu        sync.Mutex
	buckets         map[bucketKey]*bucketEntry
	idleExpiration  time.Duration
	cleanupInterval time.Duration
	nextCleanup     time.Time
	now             func() time.Time
}

func NewManager(db *gorm.DB, idleExpiration, cleanupInterval time.Duration) *Manager {
	now := time.Now
	m := &Manager{
		db: db, buckets: make(map[bucketKey]*bucketEntry), idleExpiration: idleExpiration,
		cleanupInterval: cleanupInterval, nextCleanup: now().Add(cleanupInterval), now: now,
	}
	m.snapshot.Store(&snapshot{Routes: make(map[RouteKey]routeRules)})
	return m
}

// Reload 从 MySQL 完整加载启用规则，校验成功后原子替换当前快照。
func (m *Manager) Reload() (int, error) {
	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()
	rules, err := rateLimitRepo.NewRateLimitRuleWithDB(m.db).ListEnabled()
	if err != nil {
		return 0, err
	}
	next, err := buildSnapshot(rules)
	if err != nil {
		return 0, err
	}
	m.snapshot.Store(next)
	return next.RuleCount, nil
}

func buildSnapshot(rules []rateLimitModel.RateLimitRule) (*snapshot, error) {
	next := &snapshot{Routes: make(map[RouteKey]routeRules), RuleCount: len(rules)}
	for i := range rules {
		rule := rules[i]
		if rule.RateCount <= 0 || rule.IntervalSeconds <= 0 || rule.Burst <= 0 {
			return nil, fmt.Errorf("限流规则%d的令牌参数不合法", rule.ID)
		}
		compiled := &compiledRule{
			ID: rule.ID, Limit: rate.Limit(float64(rule.RateCount) / float64(rule.IntervalSeconds)), Burst: rule.Burst,
			Fingerprint: fmt.Sprintf("%d:%d:%d", rule.RateCount, rule.IntervalSeconds, rule.Burst),
		}
		if rule.ScopeType == rateLimitModel.ScopeGlobal {
			if rule.HTTPMethod != "*" || rule.RoutePath != "*" {
				return nil, fmt.Errorf("全局限流规则%d的HTTP方法和路由必须为*", rule.ID)
			}
			if rule.Dimension == rateLimitModel.DimensionIP {
				if next.GlobalIP != nil {
					return nil, fmt.Errorf("存在重复的全局IP限流规则")
				}
				next.GlobalIP = compiled
			} else if rule.Dimension == rateLimitModel.DimensionUID {
				if next.GlobalUID != nil {
					return nil, fmt.Errorf("存在重复的全局UID限流规则")
				}
				next.GlobalUID = compiled
			} else {
				return nil, fmt.Errorf("限流规则%d的维度不合法", rule.ID)
			}
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(rule.HTTPMethod))
		if rule.ScopeType != rateLimitModel.ScopeRoute ||
			!isSupportedMethod(method) ||
			strings.TrimSpace(rule.RoutePath) != rule.RoutePath ||
			!strings.HasPrefix(rule.RoutePath, "/api/") ||
			strings.ContainsAny(rule.RoutePath, "?#") {
			return nil, fmt.Errorf("限流规则%d的作用范围或路由不合法", rule.ID)
		}
		key := RouteKey{Method: method, Path: rule.RoutePath}
		route := next.Routes[key]
		if rule.Dimension == rateLimitModel.DimensionIP {
			if route.IP != nil {
				return nil, fmt.Errorf("接口%s %s存在重复IP规则", key.Method, key.Path)
			}
			route.IP = compiled
		} else if rule.Dimension == rateLimitModel.DimensionUID {
			if route.UID != nil {
				return nil, fmt.Errorf("接口%s %s存在重复UID规则", key.Method, key.Path)
			}
			route.UID = compiled
		} else {
			return nil, fmt.Errorf("限流规则%d的维度不合法", rule.ID)
		}
		next.Routes[key] = route
	}
	return next, nil
}

// isSupportedMethod 限制第一版接口规则只使用项目当前支持的常见 HTTP 方法。
func isSupportedMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}

func (m *Manager) AllowIP(method, path, clientIP string) Decision {
	return m.allow(m.selectRule(method, path, rateLimitModel.DimensionIP), "ip:"+clientIP)
}

func (m *Manager) AllowUID(method, path string, userID int) Decision {
	return m.allow(m.selectRule(method, path, rateLimitModel.DimensionUID), "uid:"+strconv.Itoa(userID))
}

// selectRule 实现接口规则覆盖全局规则：先查接口规则，没有时才回退同维度全局规则。
func (m *Manager) selectRule(method, path string, dimension int) *compiledRule {
	current := m.snapshot.Load()
	if current == nil {
		return nil
	}
	if route, exists := current.Routes[RouteKey{Method: strings.ToUpper(method), Path: path}]; exists {
		if dimension == rateLimitModel.DimensionIP && route.IP != nil {
			return route.IP
		}
		if dimension == rateLimitModel.DimensionUID && route.UID != nil {
			return route.UID
		}
	}
	if dimension == rateLimitModel.DimensionIP {
		return current.GlobalIP
	}
	return current.GlobalUID
}

func (m *Manager) allow(rule *compiledRule, subject string) Decision {
	if rule == nil {
		return Decision{Allowed: true}
	}
	now := m.now()
	key := bucketKey{RuleID: rule.ID, Subject: subject}
	m.bucketMu.Lock()
	if !now.Before(m.nextCleanup) {
		m.cleanup(now)
		m.nextCleanup = now.Add(m.cleanupInterval)
	}
	entry := m.buckets[key]
	if entry == nil {
		entry = &bucketEntry{limiter: rate.NewLimiter(rule.Limit, rule.Burst), fingerprint: rule.Fingerprint}
		m.buckets[key] = entry
	} else if entry.fingerprint != rule.Fingerprint {
		entry.limiter.SetLimitAt(now, rule.Limit)
		entry.limiter.SetBurstAt(now, rule.Burst)
		entry.fingerprint = rule.Fingerprint
	}
	entry.lastSeen = now
	allowed := entry.limiter.AllowN(now, 1)
	retryAfter := time.Duration(0)
	if !allowed && rule.Limit > 0 {
		tokens := entry.limiter.TokensAt(now)
		retryAfter = time.Duration((1 - tokens) / float64(rule.Limit) * float64(time.Second))
		if retryAfter < time.Second {
			retryAfter = time.Second
		}
	}
	m.bucketMu.Unlock()
	return Decision{Allowed: allowed, RetryAfter: retryAfter, RuleID: rule.ID}
}

func (m *Manager) cleanup(now time.Time) {
	for key, entry := range m.buckets {
		if now.Sub(entry.lastSeen) >= m.idleExpiration {
			delete(m.buckets, key)
		}
	}
}

func (m *Manager) RuleCount() int {
	current := m.snapshot.Load()
	if current == nil {
		return 0
	}
	return current.RuleCount
}

var defaultManager atomic.Pointer[Manager]

func InitDefaultManager(db *gorm.DB, idleExpiration, cleanupInterval time.Duration) (*Manager, error) {
	manager := NewManager(db, idleExpiration, cleanupInterval)
	if _, err := manager.Reload(); err != nil {
		return nil, err
	}
	defaultManager.Store(manager)
	return manager, nil
}

func DefaultManager() *Manager { return defaultManager.Load() }

func ReloadDefault() (int, error) {
	manager := DefaultManager()
	if manager == nil {
		return 0, fmt.Errorf("默认限流管理器未初始化")
	}
	return manager.Reload()
}
