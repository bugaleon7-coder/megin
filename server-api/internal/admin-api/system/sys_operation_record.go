package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysOperationRecord @Tag 操作记录管理
type SysOperationRecord struct{}

// @Summary 获取操作记录列表
// @Description 分页查询操作记录
func (this *SysOperationRecord) GetSysOperationRecordInfoList(ctx *api.Context, req *systemDto.OperationRecordSearchReq) (*api.Result[systemDto.PageResult[systemDto.SysOperationRecord]], error) {
	result, err := systemService.NewSysOperationRecord(ctx).GetSysOperationRecordInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 删除操作记录
// @Description 根据ID删除操作记录
func (this *SysOperationRecord) DeleteSysOperationRecord(ctx *api.Context, req *systemDto.DeleteOperationRecordReq) (*api.Result[any], error) {
	err := systemService.NewSysOperationRecord(ctx).DeleteSysOperationRecord(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 批量删除操作记录
// @Description 批量删除操作记录
func (this *SysOperationRecord) DeleteSysOperationRecords(ctx *api.Context, req *DeleteIdsReq) (*api.Result[any], error) {
	err := systemService.NewSysOperationRecord(ctx).DeleteSysOperationRecords(req.Ids)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// DeleteIdsReq 批量删除请求
type DeleteIdsReq struct {
	Ids []int `json:"ids" form:"IDs[]" binding:"required,min=1"`
}
