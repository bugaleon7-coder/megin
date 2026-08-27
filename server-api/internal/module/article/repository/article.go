// 该文件由tools/gorm_gen.go生成

package repo

import (
	"megin/internal/base"
	"megin/internal/module/article/dto"
	"megin/internal/module/article/model"
	"megin/pkg/context/api"
	"time"
)

// 文章表
type Article struct {
	base.Repository[model.Article]
}

func NewArticle(ctx *api.Context) *Article {
	repo := &Article{}
	repo.Initialize(ctx)
	return repo
}

// 更新文章
func (this *Article) Update(article *model.Article) error {
	return this.DB().Save(article).Error
}

// 删除文章
func (this *Article) Delete(id int) error {
	return this.DB().Where("id", id).Delete(&model.Article{}).Error
}

// 分页查询文章列表
func (this *Article) GetPageList(req *dto.ArticleList) (*dto.PageResult[model.Article], error) {
	query := this.DB().Model(&model.Article{})

	// 添加查询条件
	if req.Category1 != "" {
		query = query.Where("category1 = ?", req.Category1)
	}
	if req.Category2 != "" {
		query = query.Where("category2 = ?", req.Category2)
	}
	if req.Source > 0 {
		query = query.Where("source = ?", req.Source)
	}
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Title != "" {
		query = query.Where("title LIKE ?", "%"+req.Title+"%")
	}
	if req.StartTime != "" {
		if startTime, err := time.Parse("2006-01-02", req.StartTime); err == nil {
			query = query.Where("create_time >= ?", startTime.Unix())
		}
	}
	if req.EndTime != "" {
		if endTime, err := time.Parse("2006-01-02", req.EndTime); err == nil {
			// 设置为当天的最后一秒
			endTime = endTime.Add(24*time.Hour - time.Second)
			query = query.Where("create_time <= ?", endTime.Unix())
		}
	}

	// 排序
	if req.SortField != "" {
		order := req.SortField
		if req.SortOrder == "" {
			req.SortOrder = "desc"
		}
		query = query.Order(order + " " + req.SortOrder)
	} else {
		query = query.Order("id desc")
	}

	// 使用通用分页查询函数
	return base.PageQuery[model.Article](query, req.PageQuery)
}
