// 该文件由tools/gorm_gen.go生成

package service

import (
	"megin/internal/base"
	"megin/internal/module/article/convert"
	"megin/internal/module/article/dto"
	"megin/internal/module/article/model"
	repo "megin/internal/module/article/repository"
	"megin/pkg/context/api"
	"time"
)

// 文章表
type Article struct {
	base.Service
	Repo *repo.Article
}

func NewArticle(ctx *api.Context) *Article {
	service := Article{}
	service.Initialize(ctx)
	service.Repo = repo.NewArticle(ctx)
	return &service
}

// 根据ID查询详情
func (this *Article) GetById(id int) (dto.Article, error) {
	// 1. 获取文章信息

	article, err := this.Repo.GetById(id)
	if err != nil {
		return dto.Article{}, err
	}
	if article.ID == 0 {
		return dto.Article{}, nil
	}
	return convert.ToArticleDTO(article), nil
}

// 创建文章
func (this *Article) Create(req *dto.CreateArticle) (model.Article, error) {
	now := time.Now().Unix()
	// 设置默认文本类型为富文本(1)
	textType := 1
	if req.TextType == 2 {
		textType = 2
	}
	article := model.Article{
		TextType:    textType,
		Content:     req.Content,
		OrigContent: req.OrigContent,
		Title:       req.Title,
		OrgTitle:    req.OrgTitle,
		Summary:     req.Summary,
		Category1:   req.Category1,
		Category2:   req.Category2,
		Source:      req.Source,
		OrigLink:    req.OrigLink,
		CreatedAt:   int(now),
		UpdatedAt:   int(now),
	}
	err := this.Repo.Create(&article)
	return article, err
}

// 更新文章
func (this *Article) Update(req *dto.UpdateArticle) (dto.Article, error) {
	// 先查询是否存在

	article, err := this.Repo.GetById(req.ID)
	if err != nil {
		return dto.Article{}, err
	}

	article.UpdatedAt = int(time.Now().Unix())
	err = this.Repo.Save(&article)
	if err != nil {
		return dto.Article{}, err
	}
	return convert.ToArticleDTO(article), nil
}

// 删除文章
func (this *Article) Delete(id int) error {
	return this.Repo.DeleteById(id)
}

// 分页查询文章列表
func (this *Article) GetPageList(req *dto.ArticleList) (*dto.PageResult[dto.Article], error) {
	// 1. 获取分页文章列表
	result, err := this.Repo.GetPageList(req)
	if err != nil {
		return nil, err
	}

	// 2. 创建结果集
	pageResult := &dto.PageResult[dto.Article]{
		PageNo:    result.PageNo,
		PageSize:  result.PageSize,
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		List:      make([]dto.Article, len(result.List)),
	}

	// 5. 处理文章数据，添加分类信息
	for i, article := range result.List {
		pageResult.List[i] = convert.ToArticleDTO(article)
	}
	return pageResult, nil
}
