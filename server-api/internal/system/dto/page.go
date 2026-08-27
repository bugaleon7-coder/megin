package dto

// PageQuery matches the pagination contract used by the original system module.
type PageQuery struct {
	PageNo   int `form:"page" json:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" json:"pageSize" binding:"required,min=1"`
}

// PageResult is the list envelope expected by the system-management frontend.
type PageResult[T any] struct {
	PageNo    int   `json:"page"`
	PageSize  int   `json:"pageSize"`
	TotalSize int64 `json:"total"`
	List      []T   `json:"list"`
}
