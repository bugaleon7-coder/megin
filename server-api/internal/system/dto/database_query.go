package dto

// DatabaseQueryReq 数据库只读查询请求。
type DatabaseQueryReq struct {
	SQL string `json:"sql" binding:"required,max=20000"`
}

// DatabaseTableReq 表结构查询请求。
type DatabaseTableReq struct {
	TableName string `json:"table_name" form:"table_name" binding:"required,max=64"`
}

// DatabaseTable 当前数据库中的数据表。
type DatabaseTable struct {
	Name        string `json:"name"`
	Comment     string `json:"comment"`
	Rows        int64  `json:"rows"`
	OrderColumn string `json:"order_column"`
}

// DatabaseTableStructure 数据库返回的原生建表语句。
type DatabaseTableStructure struct {
	TableName string `json:"table_name"`
	CreateSQL string `json:"create_sql"`
}

// DatabaseQueryColumn 查询结果字段信息。
type DatabaseQueryColumn struct {
	Name         string `json:"name"`
	DatabaseType string `json:"database_type"`
}

// DatabaseQueryResult 数据库查询结果。
type DatabaseQueryResult struct {
	Columns    []DatabaseQueryColumn `json:"columns"`
	Rows       [][]any               `json:"rows"`
	RowCount   int                   `json:"row_count"`
	DurationMS int64                 `json:"duration_ms"`
	Truncated  bool                  `json:"truncated"`
}
