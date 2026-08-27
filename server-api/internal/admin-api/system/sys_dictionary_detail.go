package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysDictionaryDetail @Tag 字典详情管理
type SysDictionaryDetail struct{}

// @Summary 创建字典详情
// @Description 创建字典详情项
func (this *SysDictionaryDetail) CreateSysDictionaryDetail(ctx *api.Context, req *systemDto.CreateDictionaryDetailReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionaryDetail(ctx).CreateSysDictionaryDetail(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除字典详情
// @Description 删除字典详情项
func (this *SysDictionaryDetail) DeleteSysDictionaryDetail(ctx *api.Context, req *systemDto.DeleteDictionaryDetailReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionaryDetail(ctx).DeleteSysDictionaryDetail(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新字典详情
// @Description 更新字典详情项信息
func (this *SysDictionaryDetail) UpdateSysDictionaryDetail(ctx *api.Context, req *systemDto.UpdateDictionaryDetailReq) (*api.Result[any], error) {
	err := systemService.NewSysDictionaryDetail(ctx).UpdateSysDictionaryDetail(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取字典详情
// @Description 根据ID获取字典详情信息
func (this *SysDictionaryDetail) GetSysDictionaryDetail(ctx *api.Context, req *systemDto.GetDictionaryDetailByIdReq) (*api.Result[systemDto.SysDictionaryDetail], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetSysDictionaryDetail(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 查询字典详情
// @Description 根据ID查询字典详情信息
func (this *SysDictionaryDetail) FindSysDictionaryDetail(ctx *api.Context, req *systemDto.GetDictionaryDetailByIdReq) (*api.Result[systemDto.DictionaryDetailResponse], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetSysDictionaryDetail(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.DictionaryDetailResponse{DictionaryDetail: *result})
}

// @Summary 获取字典详情列表
// @Description 分页查询字典详情列表
func (this *SysDictionaryDetail) GetSysDictionaryDetailInfoList(ctx *api.Context, req *systemDto.DictionaryDetailSearchReq) (*api.Result[systemDto.PageResult[systemDto.SysDictionaryDetail]], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetSysDictionaryDetailInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 获取字典详情列表
// @Description 按前端契约分页查询字典详情列表
func (this *SysDictionaryDetail) GetSysDictionaryDetailList(ctx *api.Context, req *systemDto.DictionaryDetailSearchReq) (*api.Result[systemDto.PageResult[systemDto.SysDictionaryDetail]], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetSysDictionaryDetailInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 根据字典ID获取详情
// @Description 根据字典ID获取详情树形结构
func (this *SysDictionaryDetail) GetDetailsByDictionaryID(ctx *api.Context, req *GetDictDetailsReq) (*api.Result[[]systemDto.SysDictionaryDetail], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetDetailsByDictionaryID(req.DictID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 获取字典详情树
// @Description 根据字典ID获取树形结构
func (this *SysDictionaryDetail) GetDictionaryTreeList(ctx *api.Context, req *systemDto.GetDictionaryTreeReq) (*api.Result[systemDto.DictionaryDetailListResponse], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetDictionaryTreeList(req.SysDictionaryID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.DictionaryDetailListResponse{List: result})
}

// @Summary 根据字典类型获取字典详情树
// @Description 根据字典类型获取树形结构
func (this *SysDictionaryDetail) GetDictionaryTreeListByType(ctx *api.Context, req *systemDto.GetDictionaryTreeByTypeReq) (*api.Result[systemDto.DictionaryDetailListResponse], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetDictionaryTreeListByType(req.Type)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.DictionaryDetailListResponse{List: result})
}

// @Summary 根据父ID获取详情
// @Description 根据父ID获取字典详情列表
func (this *SysDictionaryDetail) GetDetailsByParent(ctx *api.Context, req *systemDto.GetDetailsByParentReq) (*api.Result[[]systemDto.SysDictionaryDetail], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetDetailsByParent(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 根据父ID获取详情
// @Description 根据父ID获取字典详情列表
func (this *SysDictionaryDetail) GetDictionaryDetailsByParent(ctx *api.Context, req *systemDto.GetDetailsByParentReq) (*api.Result[systemDto.DictionaryDetailListResponse], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetDetailsByParent(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.DictionaryDetailListResponse{List: result})
}

// @Summary 获取字典详情路径
// @Description 获取字典详情的完整路径
func (this *SysDictionaryDetail) GetDictionaryPath(ctx *api.Context, req *systemDto.GetDictionaryPathReq) (*api.Result[systemDto.DictionaryDetailListResponse], error) {
	result, err := systemService.NewSysDictionaryDetail(ctx).GetDictionaryPath(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.DictionaryDetailListResponse{List: result})
}

// GetDictDetailsReq 根据字典ID查询详情请求
type GetDictDetailsReq struct {
	DictID int `json:"dictID" form:"dictID" binding:"required"`
}
