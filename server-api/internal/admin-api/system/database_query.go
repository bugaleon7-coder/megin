package system

import (
	commonDto "megin/internal/dto"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// DatabaseQuery @Tag 数据库查询
type DatabaseQuery struct{}

// Tables @Summary 获取当前数据库的数据表
// @Description 返回当前数据库中的全部基础表和视图，供数据库查询页面下拉选择。
func (h *DatabaseQuery) Tables(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[[]systemDto.DatabaseTable], error) {
	result, err := systemService.NewDatabaseQuery(ctx).Tables()
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// TableStructure @Summary 获取数据表字段结构
// @Description 根据当前数据库和表名返回 SHOW CREATE TABLE 或 SHOW CREATE VIEW 的原生文本。
func (h *DatabaseQuery) TableStructure(ctx *api.Context, req *systemDto.DatabaseTableReq) (*api.Result[systemDto.DatabaseTableStructure], error) {
	result, err := systemService.NewDatabaseQuery(ctx).TableStructure(req.TableName)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// Execute @Summary 执行只读数据库查询
// @Description 只允许执行一条以 SELECT 开始并包含最外层 LIMIT 的查询；禁止所有写入、DDL、锁和导出操作。
func (h *DatabaseQuery) Execute(ctx *api.Context, req *systemDto.DatabaseQueryReq) (*api.Result[systemDto.DatabaseQueryResult], error) {
	result, err := systemService.NewDatabaseQuery(ctx).Execute(req.SQL)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}
