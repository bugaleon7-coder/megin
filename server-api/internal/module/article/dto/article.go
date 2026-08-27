package dto

import commonDto "megin/internal/module/common/dto"

type PageQuery = commonDto.PageQuery
type PageResult[T any] = commonDto.PageResult[T]

// Article 文章信息响应结构体
type Article struct {
	ID        int    `json:"id" form:"id"` // 文章ID
	TextType  int    `json:"text_type"`    // 文本类型 1 富文本 2 纯文本
	Content   string `json:"content"`      // 文章内容
	Title     string `json:"title"`        // 文章标题
	Summary   string `json:"summary"`      // 文章摘要
	Category1 int    `json:"category1"`    // 一级分类ID
	Category2 int    `json:"category2"`    // 二级分类ID
	CreatedAt int    `json:"created_at"`   // 创建时间戳
	UpdatedAt int    `json:"updated_at"`   // 更新时间戳
	Source    int    `json:"source"`       // 文章来源
	OrigLink  string `json:"orig_link"`    // 原始链接
}

// CreateArticle 创建文章请求结构体
type CreateArticle struct {
	TextType    int    `json:"text_type" binding:"omitempty,oneof=1 2" example:"1"`                                                                         // 文本类型 1 富文本 2 纯文本，默认1
	Content     string `json:"content" binding:"required,min=1,max=20000" example:"这是一篇新文章的内容..."`                                                          // 文章内容
	OrigContent string `json:"orig_content" binding:"omitempty,min=1,max=50000" example:"原始文章的完整内容..."`                                                     // 原始文章内容
	Title       string `json:"title" binding:"required,min=1,max=200" example:"新文章标题"`                                                                      // 文章标题
	OrgTitle    string `json:"org_title" binding:"omitempty,min=1,max=200" example:"原始文章标题"`                                                                // 原始文章标题
	Summary     string `json:"summary" binding:"omitempty,min=1,max=500" example:"这是文章的摘要信息..."`                                                            // 文章摘要
	Category1   int    `json:"category1" binding:"min=1" example:"1"`                                                                                       // 一级分类ID (min=1已经包含了required的效果)
	Category2   int    `json:"category2" binding:"omitempty,min=0" example:"0"`                                                                             // 二级分类ID
	Source      int    `json:"source" binding:"min=1" example:"1"`                                                                                          // 文章来源 (min=1已经包含了required的效果)
	OrigLink    string `json:"orig_link" binding:"omitempty,url,min=1,max=255" swaggertype:"string" format:"url" example:"https://example.com/article/123"` // 原始链接
}

// UpdateArticle 更新文章请求结构体
type UpdateArticle struct {
	ID        int    `json:"id" binding:"required,min=1"`                                                       // 文章ID
	TextType  int    `json:"text_type" binding:"omitempty,oneof=1 2"`                                           // 文本类型 1 富文本 2 纯文本
	Content   string `json:"content" binding:"omitempty,min=1,max=20000"`                                       // 文章内容
	Title     string `json:"title" binding:"omitempty,min=1,max=200"`                                           // 文章标题
	Summary   string `json:"summary" binding:"omitempty,min=1,max=500"`                                         // 文章摘要
	Category1 int    `json:"category1" binding:"omitempty,min=1"`                                               // 一级分类ID
	Category2 int    `json:"category2" binding:"omitempty,min=0"`                                               // 二级分类ID
	Source    int    `json:"source" binding:"omitempty,min=1"`                                                  // 文章来源
	OrigLink  string `json:"orig_link" binding:"omitempty,url,min=1,max=255" swaggertype:"string" format:"url"` // 原始链接
}

// ArticleList 文章列表查询请求结构体
type ArticleList struct {
	Category1 string `form:"category1" json:"category1" binding:"omitempty,min=1"`            // 一级分类ID
	Category2 string `form:"category2" json:"category2" binding:"omitempty,min=1"`            // 二级分类ID
	Source    int    `form:"source" json:"source" binding:"omitempty,min=1"`                  // 文章来源
	Status    int    `form:"status" json:"status" binding:"omitempty,oneof=0 1 2"`            // 文章状态
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=100"`              // 关键词搜索
	Title     string `form:"title" json:"title" binding:"omitempty,max=100"`                  // 标题搜索
	StartTime string `form:"start_time" json:"start_time" binding:"omitempty,date"`           // 开始时间
	EndTime   string `form:"end_time" json:"end_time" binding:"omitempty,date"`               // 结束时间
	SortField string `form:"sort_field" json:"sort_field" binding:"omitempty,max=50"`         // 排序字段
	SortOrder string `form:"sort_order" json:"sort_order" binding:"omitempty,oneof=asc desc"` // 排序方向
	PageQuery        // 嵌入通用分页查询参数
}
