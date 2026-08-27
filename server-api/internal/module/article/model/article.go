// 该文件由tools/gorm_gen.go生成

package model

const TableNameArticle = "articles"

// Article 文章表
type Article struct {
	ID          int  `gorm:"column:id;primaryKey;autoIncrement:true;comment:自增主键" json:"id"` // 自增主键
	TextType    int  `gorm:"column:text_type;not null;default:1;comment:文本类型 1 富文本 2 纯文本" json:"text_type"` // 文本类型 1 富文本 2 纯文本
	Content     string `gorm:"column:content;comment:内容" json:"content"`                       // 内容
	OrigContent string `gorm:"column:orig_content;comment:原始内容" json:"orig_content"`           // 原始内容
	Title       string `gorm:"column:title;not null;comment:标题" json:"title"`                  // 标题
	OrgTitle    string `gorm:"column:org_title;not null;comment:原始标题" json:"org_title"`        // 原始标题
	Summary     string `gorm:"column:summary;not null;comment:摘要" json:"summary"`              // 摘要
	Category1   int  `gorm:"column:category1;not null;comment:类别1" json:"category1"`         // 类别1
	Category2   int  `gorm:"column:category2;not null;comment:类型2" json:"category2"`         // 类型2
	CreatedAt   int  `gorm:"column:created_at;not null;comment:创建时间" json:"created_at"`      // 创建时间
	UpdatedAt   int  `gorm:"column:updated_at;not null;comment:修改时间" json:"updated_at"`      // 修改时间
	Source      int  `gorm:"column:source;not null;comment:来源" json:"source"`                // 来源
	OrigLink    string `gorm:"column:orig_link;not null;comment:原始链接" json:"orig_link"`        // 原始链接
}

// 计录是否存在
func (m Article) IsNil() bool {
	return m.ID == 0
}

// 获取ID
func (m Article) GetID() any {
	return m.ID
}

// TableName Article's table name
func (Article) TableName() string {
	return TableNameArticle
}
