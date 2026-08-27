package dto

import "github.com/golang-jwt/jwt/v5"

type PageResult[T any] struct {
	//上面的分页页码
	PageNo    int   `json:"page_no"`   //分页页码
	PageSize  int   `json:"page_size"` //分页每页条数
	TotalSize int64 `json:"total_size"`
	TotalPage int64 `json:"total_page"`
	List      []T   `json:"list"` //分页数据列表
}

// EmptyReq 用于不接收任何请求参数的接口。
type EmptyReq struct{}

// PageQuery 通用分页查询请求结构体
type PageQuery struct {
	PageNo   int `form:"page_no" json:"page_no" binding:"required,min=1"`             //分页页码
	PageSize int `form:"page_size" json:"page_size" binding:"required,min=1,max=100"` //每页条数
}

// BaseID 通用ID结构体
type BaseID[T any] struct {
	ID T `json:"id"` // ID是一个可变类型的参数
}

// BaseQueryByIdReq 通用按ID查询请求结构体
type BaseQueryByIdReq struct {
	ID int `json:"id" form:"id" binding:"required,min=1"` // 记录ID
}

// BaseDeleteByIdReq 通用按ID删除请求结构体
type BaseDeleteByIdReq struct {
	ID int `json:"id" form:"id" binding:"required,min=1"` // 记录ID
}

// BaseModifyStatusByIdReq 通用修改状态请求结构体
type BaseModifyStatusByIdReq struct {
	ID     int `json:"id" form:"id" binding:"required,min=1"`      // 记录ID
	Status int `json:"status" form:"status" binding:"oneof=0 1 2"` // 状态值
}

// BaseQueryPageListReq 通用分页查询请求结构体
type BaseQueryPageListReq struct {
	PageQuery     // 嵌入通用分页查询参数
	Status    int `json:"status" form:"status" binding:"oneof=0 1 2"` // 状态值
}

// QueryFilter 通用查询过滤器
type QueryFilter struct {
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=100"`              // 关键词搜索
	StartTime string `form:"start_time" json:"start_time" binding:"omitempty,date"`           // 开始时间
	EndTime   string `form:"end_time" json:"end_time" binding:"omitempty,date"`               // 结束时间
	SortField string `form:"sort_field" json:"sort_field" binding:"omitempty,max=50"`         // 排序字段
	SortOrder string `form:"sort_order" json:"sort_order" binding:"omitempty,oneof=asc desc"` // 排序方向
}

// AdvancedQueryPageListReq 高级分页查询请求结构体
type AdvancedQueryPageListReq struct {
	PageQuery   // 嵌入通用分页查询参数
	QueryFilter // 嵌入通用查询过滤器
}

const (
	ApiClaimToken = "ApiClaimToken"
	ApiJwtClaims  = "ApiJwtClaims"
)

const (
	AdminApiClaimToken = "AdminApiClaimToken"
	AdminApiJwtClaims  = "AdminApiJwtClaims"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	RoleId   int    `json:"role_id" comment:"角色ID"`
	jwt.RegisteredClaims
}
