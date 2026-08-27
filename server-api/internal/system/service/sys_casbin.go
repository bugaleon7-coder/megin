package service

import (
	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"megin/internal/base"
	"megin/internal/config"
	systemDto "megin/internal/system/dto"
	systemModel "megin/internal/system/model"
	"megin/pkg/context/api"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm"
)

type SysCasbin struct {
	base.Service
}

const adminAPIRouterPrefix = "/admin-api"

var (
	casbinEnforcerMu       sync.RWMutex
	casbinEnforcerInitOnce sync.Once
	casbinEnforcerInstance *casbin.Enforcer
	casbinEnforcerInitErr  error
	casbinDBProvider       = config.GetMysqlDB
)

func NewSysCasbin(ctx *api.Context) *SysCasbin {
	s := &SysCasbin{}
	s.Initialize(ctx)
	return s
}

func (s *SysCasbin) getEnforcer() (*casbin.Enforcer, error) {
	casbinEnforcerInitOnce.Do(func() {
		casbinEnforcerInstance, casbinEnforcerInitErr = newCasbinEnforcer(casbinDBProvider())
	})
	if casbinEnforcerInitErr != nil {
		return nil, casbinEnforcerInitErr
	}
	casbinEnforcerMu.RLock()
	defer casbinEnforcerMu.RUnlock()
	return casbinEnforcerInstance, nil
}

func (s *SysCasbin) UpdateCasbin(authorityId uint, casbinInfos []systemDto.CasbinInfo) error {
	e, err := s.getEnforcer()
	if err != nil {
		return s.Error(err, "获取Casbin Enforcer失败")
	}
	casbinEnforcerMu.Lock()
	defer casbinEnforcerMu.Unlock()

	// Remove old policies
	_, err = e.RemoveFilteredPolicy(0, s.authorityIdStr(authorityId))
	if err != nil {
		return s.Error(err, "清除旧策略失败")
	}

	// Add new policies
	for _, ci := range casbinInfos {
		_, err = e.AddPolicy(s.authorityIdStr(authorityId), NormalizeCasbinPath(ci.Path), ci.Method)
		if err != nil {
			return s.Error(err, "添加策略失败")
		}
	}
	return e.LoadPolicy()
}

func (s *SysCasbin) GetPolicyPathByAuthorityId(authorityId uint) ([]systemDto.CasbinInfo, error) {
	e, err := s.getEnforcer()
	if err != nil {
		return nil, s.Error(err, "获取Casbin Enforcer失败")
	}
	casbinEnforcerMu.RLock()
	defer casbinEnforcerMu.RUnlock()

	policies, err := e.GetFilteredPolicy(0, s.authorityIdStr(authorityId))
	if err != nil {
		return nil, s.Error(err, "查询策略失败")
	}
	var casbinInfos []systemDto.CasbinInfo
	for _, p := range policies {
		if len(p) >= 3 {
			casbinInfos = append(casbinInfos, systemDto.CasbinInfo{
				Path:   p[1],
				Method: p[2],
			})
		}
	}
	return casbinInfos, nil
}

func (s *SysCasbin) ClearCasbin(v0 string, v1 string, v2 string) error {
	e, err := s.getEnforcer()
	if err != nil {
		return s.Error(err, "获取Casbin Enforcer失败")
	}
	casbinEnforcerMu.Lock()
	defer casbinEnforcerMu.Unlock()

	_, err = e.RemoveFilteredPolicy(0, v0, v1, v2)
	if err != nil {
		return err
	}
	return e.LoadPolicy()
}

func (s *SysCasbin) FreshCasbin(authorityId uint) error {
	e, err := s.getEnforcer()
	if err != nil {
		return s.Error(err, "获取Casbin Enforcer失败")
	}
	casbinEnforcerMu.Lock()
	defer casbinEnforcerMu.Unlock()
	return e.LoadPolicy()
}

// GetAuthoritiesByApi 获取拥有指定API权限的所有角色ID
func (s *SysCasbin) GetAuthoritiesByApi(path, method string) ([]uint, error) {
	var rules []CasbinRule
	path = NormalizeCasbinPath(path)
	err := config.GetMysqlDB().Model(&CasbinRule{}).
		Where("ptype = 'p' AND v1 = ? AND v2 = ?", path, method).
		Find(&rules).Error
	if err != nil {
		return nil, s.Error(err, "查询API角色失败")
	}
	authorityIds := make([]uint, 0, len(rules))
	for _, r := range rules {
		id, e := strconv.Atoi(r.V0)
		if e == nil {
			authorityIds = append(authorityIds, uint(id))
		}
	}
	return authorityIds, nil
}

// SetApiAuthorities 全量覆盖某API关联的角色列表
func (s *SysCasbin) SetApiAuthorities(path, method string, authorityIds []uint) error {
	path = NormalizeCasbinPath(path)
	if err := casbinDBProvider().Transaction(func(tx *gorm.DB) error {
		// 1. Delete all existing role associations for this API
		if err := tx.Where("ptype = 'p' AND v1 = ? AND v2 = ?", path, method).
			Delete(&CasbinRule{}).Error; err != nil {
			return err
		}
		// 2. Batch insert new associations
		if len(authorityIds) > 0 {
			newRules := make([]CasbinRule, 0, len(authorityIds))
			for _, authorityId := range authorityIds {
				newRules = append(newRules, CasbinRule{
					PType: "p",
					V0:    strconv.Itoa(int(authorityId)),
					V1:    path,
					V2:    method,
				})
			}
			if err := tx.Create(&newRules).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return s.FreshCasbin(0)
}

func (s *SysCasbin) UpdateCasbinApi(oldPath, newPath, oldMethod, newMethod string) error {
	e, err := s.getEnforcer()
	if err != nil {
		return s.Error(err, "获取Casbin Enforcer失败")
	}
	oldPath = NormalizeCasbinPath(oldPath)
	newPath = NormalizeCasbinPath(newPath)
	casbinEnforcerMu.Lock()
	defer casbinEnforcerMu.Unlock()

	// Update all policies with old path/method to new ones
	policies, err := e.GetPolicy()
	if err != nil {
		return s.Error(err, "查询策略失败")
	}
	for _, p := range policies {
		if len(p) >= 3 && p[1] == oldPath && p[2] == oldMethod {
			_, err = e.RemovePolicy(p[0], oldPath, oldMethod)
			if err != nil {
				return err
			}
			_, err = e.AddPolicy(p[0], newPath, newMethod)
			if err != nil {
				return err
			}
		}
	}
	return e.LoadPolicy()
}

func (s *SysCasbin) authorityIdStr(id uint) string {
	return s.formatID(id)
}

func EnforceAuthorityPolicy(authorityId uint, path, method string) (bool, error) {
	service := &SysCasbin{}
	enforcer, err := service.getEnforcer()
	if err != nil {
		return false, err
	}
	casbinEnforcerMu.RLock()
	defer casbinEnforcerMu.RUnlock()
	return enforcer.Enforce(service.formatID(authorityId), NormalizeCasbinPath(path), method)
}

func IsDefaultAuthenticatedPolicy(path, method string) bool {
	normalizedPath := NormalizeCasbinPath(path)
	for _, policy := range systemDto.DefaultCasbinInfos {
		if policy.Path == normalizedPath && strings.EqualFold(policy.Method, method) {
			return true
		}
	}
	return false
}

func (s *SysCasbin) formatID(id uint) string {
	return strconv.Itoa(int(id))
}

func NormalizeCasbinPath(path string) string {
	return strings.TrimPrefix(path, adminAPIRouterPrefix)
}

func newCasbinEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}
	m, err := casbinmodel.NewModelFromString(casbinModelText)
	if err != nil {
		return nil, err
	}
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	if err = enforcer.LoadPolicy(); err != nil {
		return nil, err
	}
	return enforcer, nil
}

const casbinModelText = `
	[request_definition]
	r = sub, obj, act

	[policy_definition]
	p = sub, obj, act

	[role_definition]
	g = _, _

	[policy_effect]
	e = some(where (p.eft == allow))

	[matchers]
	m = r.sub == p.sub && keyMatch2(r.obj,p.obj) && r.act == p.act
`

// CasbinRule is used to directly access the casbin_rule table
type CasbinRule struct {
	PType string `gorm:"column:ptype"`
	V0    string `gorm:"column:v0"`
	V1    string `gorm:"column:v1"`
	V2    string `gorm:"column:v2"`
	V3    string `gorm:"column:v3"`
	V4    string `gorm:"column:v4"`
	V5    string `gorm:"column:v5"`
}

func (CasbinRule) TableName() string {
	return "casbin_rule"
}

// Ensure model.Model interface requirement
var _ interface{ TableName() string } = (*CasbinRule)(nil)
var _ interface{ TableName() string } = (*systemModel.SysUser)(nil)
