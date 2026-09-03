package dto

import "github.com/golang-jwt/jwt/v5"

type PageResult[T any] struct {
	PageNo    int   `json:"page_no"`    // 当前页码。
	PageSize  int   `json:"page_size"`  // 每页记录数。
	TotalSize int64 `json:"total_size"` // 总记录数。
	TotalPage int64 `json:"total_page"` // 总页数。
	List      []T   `json:"list"`       // 当前页数据列表。
}

type EmptyReq struct{}

type PageQuery struct {
	PageNo   int `form:"page_no" json:"page_no" binding:"required,min=1"`             // 页码，从 1 开始。
	PageSize int `form:"page_size" json:"page_size" binding:"required,min=1,max=100"` // 每页条数，最大 100。
}

type BaseID[T any] struct {
	ID T `json:"id"` // 业务对象 ID。
}
type BaseQueryByIdReq struct {
	ID int `json:"id" form:"id" binding:"required,min=1"` // 要查询的对象 ID。
}
type BaseDeleteByIdReq struct {
	ID int `json:"id" form:"id" binding:"required,min=1"` // 要删除的对象 ID。
}
type BaseModifyStatusByIdReq struct {
	ID     int `json:"id" form:"id" binding:"required,min=1"`      // 要修改状态的对象 ID。
	Status int `json:"status" form:"status" binding:"oneof=0 1 2"` // 新状态：0、1 或 2。
}
type BaseQueryPageListReq struct {
	PageQuery     // 通用分页参数。
	Status    int `json:"status" form:"status" binding:"oneof=0 1 2"` // 状态筛选条件：0、1 或 2。
}
type QueryFilter struct {
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=100"`              // 关键字筛选条件。
	StartTime string `form:"start_time" json:"start_time" binding:"omitempty,date"`           // 日期范围起点，格式 YYYY-MM-DD。
	EndTime   string `form:"end_time" json:"end_time" binding:"omitempty,date"`               // 日期范围终点，格式 YYYY-MM-DD。
	SortField string `form:"sort_field" json:"sort_field" binding:"omitempty,max=50"`         // 排序字段。
	SortOrder string `form:"sort_order" json:"sort_order" binding:"omitempty,oneof=asc desc"` // 排序方向：asc 或 desc。
}
type AdvancedQueryPageListReq struct {
	PageQuery   // 通用分页参数。
	QueryFilter // 通用筛选与排序参数。
}

const (
	ApiClaimToken      = "ApiClaimToken"
	ApiJwtClaims       = "ApiJwtClaims"
	AdminApiClaimToken = "AdminApiClaimToken"
	AdminApiJwtClaims  = "AdminApiJwtClaims"
)

type Claims struct {
	UserID               int    `json:"user_id"`                // 用户 ID。
	Username             string `json:"username"`               // 用户名。
	Mobile               string `json:"mobile"`                 // 用户手机号。
	RoleId               int    `json:"role_id" comment:"角色ID"` // 角色 ID。
	jwt.RegisteredClaims        // JWT 标准注册声明（如过期时间）。
}
