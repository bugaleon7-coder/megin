package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysCasbin @Tag Casbin权限管理
type SysCasbin struct{}

// @Summary 更新Casbin策略
// @Description 更新角色的Casbin权限策略
func (this *SysCasbin) UpdateCasbin(ctx *api.Context, req *systemDto.CasbinCreateReq) (*api.Result[any], error) {
	err := systemService.NewSysCasbin(ctx).UpdateCasbin(req.AuthorityId, req.CasbinInfos)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取策略路径
// @Description 根据角色ID获取Casbin策略路径
func (this *SysCasbin) GetPolicyPathByAuthorityId(ctx *api.Context, req *systemDto.AuthorityIdReq) (*api.Result[systemDto.PolicyPathResponse], error) {
	result, err := systemService.NewSysCasbin(ctx).GetPolicyPathByAuthorityId(req.AuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.PolicyPathResponse{Paths: result})
}

// @Summary 清除Casbin策略
// @Description 清除指定角色的Casbin策略
func (this *SysCasbin) ClearCasbin(ctx *api.Context, req *systemDto.AuthorityIdReq) (*api.Result[any], error) {
	err := systemService.NewSysCasbin(ctx).ClearCasbin("", "", "")
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 刷新Casbin策略
// @Description 从数据库重新加载Casbin策略
func (this *SysCasbin) FreshCasbin(ctx *api.Context, req *systemDto.AuthorityIdReq) (*api.Result[any], error) {
	err := systemService.NewSysCasbin(ctx).FreshCasbin(req.AuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新Casbin API
// @Description 更新Casbin中所有匹配的API路径和方法
func (this *SysCasbin) UpdateCasbinApi(ctx *api.Context, req *CasbinApiUpdateReq) (*api.Result[any], error) {
	err := systemService.NewSysCasbin(ctx).UpdateCasbinApi(req.OldPath, req.NewPath, req.OldMethod, req.NewMethod)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// CasbinApiUpdateReq Casbin API更新请求
type CasbinApiUpdateReq struct {
	OldPath   string `json:"oldPath" binding:"required"`
	NewPath   string `json:"newPath" binding:"required"`
	OldMethod string `json:"oldMethod" binding:"required"`
	NewMethod string `json:"newMethod" binding:"required"`
}
