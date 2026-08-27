package system

import (
	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysDictionary @Tag 字典管理
type SysDictionary struct{}

// @Summary 创建字典
// @Description 创建系统字典
func (this *SysDictionary) CreateSysDictionary(ctx *api.Context, req *systemDto.CreateDictionaryReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionary(ctx).CreateSysDictionary(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除字典
// @Description 删除系统字典
func (this *SysDictionary) DeleteSysDictionary(ctx *api.Context, req *systemDto.DeleteDictionaryReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionary(ctx).DeleteSysDictionary(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新字典
// @Description 更新系统字典信息
func (this *SysDictionary) UpdateSysDictionary(ctx *api.Context, req *systemDto.UpdateDictionaryReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionary(ctx).UpdateSysDictionary(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取字典详情
// @Description 根据ID获取字典详细信息
func (this *SysDictionary) GetSysDictionary(ctx *api.Context, req *systemDto.GetDictionaryByIdReq) (*api.Result[systemDto.SysDictionary], error) {
	result, err := systemService.NewSysDictionary(ctx).GetSysDictionary(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 查询字典
// @Description 根据 ID 或类型查询字典
func (this *SysDictionary) FindSysDictionary(ctx *api.Context, req *systemDto.FindDictionaryReq) (*api.Result[systemDto.DictionaryResponse], error) {
	result, err := systemService.NewSysDictionary(ctx).FindSysDictionary(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.DictionaryResponse{Dictionary: *result})
}

// @Summary 获取字典列表
// @Description 分页查询字典列表
func (this *SysDictionary) GetSysDictionaryInfoList(ctx *api.Context, req *systemDto.DictionarySearchReq) (*api.Result[systemDto.PageResult[systemDto.SysDictionary]], error) {
	result, err := systemService.NewSysDictionary(ctx).GetSysDictionaryInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 获取字典树列表
// @Description 按名称或类型查询字典列表
func (this *SysDictionary) GetSysDictionaryList(ctx *api.Context, req *systemDto.DictionaryListReq) (*api.Result[[]systemDto.SysDictionary], error) {
	result, err := systemService.NewSysDictionary(ctx).GetSysDictionaryList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 导出字典
// @Description 导出所有字典数据
func (this *SysDictionary) ExportSysDictionary(ctx *api.Context, req *commonDto.BaseQueryByIdReq) (*api.Result[[]systemDto.SysDictionary], error) {
	result, err := systemService.NewSysDictionary(ctx).ExportSysDictionary()
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 导入字典
// @Description 从JSON导入字典数据
func (this *SysDictionary) ImportSysDictionary(ctx *api.Context, req *systemDto.ImportDictionaryReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionary(ctx).ImportSysDictionary(req.Json)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}
