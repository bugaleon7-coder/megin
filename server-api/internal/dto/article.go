package dto

type Article struct {
	ID        int    `json:"id" form:"id"` // 文章 ID。
	TextType  int    `json:"text_type"`    // 正文类型：1 为富文本，2 为 Markdown。
	Content   string `json:"content"`      // 当前展示的文章正文。
	Title     string `json:"title"`        // 文章标题。
	Summary   string `json:"summary"`      // 文章摘要。
	Category1 int    `json:"category1"`    // 一级分类 ID。
	Category2 int    `json:"category2"`    // 二级分类 ID，未分类时为 0。
	CreatedAt int    `json:"created_at"`   // 创建时间的 Unix 时间戳。
	UpdatedAt int    `json:"updated_at"`   // 最后更新时间的 Unix 时间戳。
	Source    int    `json:"source"`       // 文章来源类型。
	OrigLink  string `json:"orig_link"`    // 原文链接。
}
type CreateArticle struct {
	TextType    int    `json:"text_type" binding:"omitempty,oneof=1 2" example:"1"`                                                                         // 正文类型：1 富文本，2 Markdown；省略时由服务端处理。
	Content     string `json:"content" binding:"required,min=1,max=20000" example:"这是一篇新文章的内容..."`                                                          // 文章正文，必填。
	OrigContent string `json:"orig_content" binding:"omitempty,min=1,max=50000" example:"原始文章的完整内容..."`                                                     // 原始正文，用于保留导入前内容。
	Title       string `json:"title" binding:"required,min=1,max=200" example:"新文章标题"`                                                                      // 当前文章标题，必填。
	OrgTitle    string `json:"org_title" binding:"omitempty,min=1,max=200" example:"原始文章标题"`                                                                // 原始文章标题。
	Summary     string `json:"summary" binding:"omitempty,min=1,max=500" example:"这是文章的摘要信息..."`                                                            // 文章摘要。
	Category1   int    `json:"category1" binding:"min=1" example:"1"`                                                                                       // 一级分类 ID，必填。
	Category2   int    `json:"category2" binding:"omitempty,min=0" example:"0"`                                                                             // 二级分类 ID，0 表示未指定。
	Source      int    `json:"source" binding:"min=1" example:"1"`                                                                                          // 文章来源类型，必填。
	OrigLink    string `json:"orig_link" binding:"omitempty,url,min=1,max=255" swaggertype:"string" format:"url" example:"https://example.com/article/123"` // 原文链接。
}
type UpdateArticle struct {
	ID        int    `json:"id" binding:"required,min=1"`                                                       // 要更新的文章 ID。
	TextType  int    `json:"text_type" binding:"omitempty,oneof=1 2"`                                           // 正文类型：1 富文本，2 Markdown。
	Content   string `json:"content" binding:"omitempty,min=1,max=20000"`                                       // 更新后的文章正文。
	Title     string `json:"title" binding:"omitempty,min=1,max=200"`                                           // 更新后的文章标题。
	Summary   string `json:"summary" binding:"omitempty,min=1,max=500"`                                         // 更新后的文章摘要。
	Category1 int    `json:"category1" binding:"omitempty,min=1"`                                               // 更新后的一级分类 ID。
	Category2 int    `json:"category2" binding:"omitempty,min=0"`                                               // 更新后的二级分类 ID。
	Source    int    `json:"source" binding:"omitempty,min=1"`                                                  // 更新后的文章来源类型。
	OrigLink  string `json:"orig_link" binding:"omitempty,url,min=1,max=255" swaggertype:"string" format:"url"` // 更新后的原文链接。
}
type ArticleList struct {
	Category1 string `form:"category1" json:"category1" binding:"omitempty,min=1"`            // 一级分类筛选条件。
	Category2 string `form:"category2" json:"category2" binding:"omitempty,min=1"`            // 二级分类筛选条件。
	Source    int    `form:"source" json:"source" binding:"omitempty,min=1"`                  // 文章来源类型筛选条件。
	Status    int    `form:"status" json:"status" binding:"omitempty,oneof=0 1 2"`            // 文章状态筛选：0、1 或 2。
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=100"`              // 全文关键字筛选条件。
	Title     string `form:"title" json:"title" binding:"omitempty,max=100"`                  // 标题筛选条件。
	StartTime string `form:"start_time" json:"start_time" binding:"omitempty,date"`           // 创建日期起点，格式 YYYY-MM-DD。
	EndTime   string `form:"end_time" json:"end_time" binding:"omitempty,date"`               // 创建日期终点，格式 YYYY-MM-DD。
	SortField string `form:"sort_field" json:"sort_field" binding:"omitempty,max=50"`         // 排序字段。
	SortOrder string `form:"sort_order" json:"sort_order" binding:"omitempty,oneof=asc desc"` // 排序方向：asc 或 desc。
	PageQuery        // 通用分页参数。
}
