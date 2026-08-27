// 该文件由tools/gorm_gen.go生成
package handler

import (
	"megin/internal/module/article/dto"
	"megin/internal/module/article/service"
	"megin/pkg/context/api"
)

// Article @Tag 文章管理模块
type Article struct {
}

// Detail @Summary 根据ID查询文章详情
// @Description 根据文章ID获取文章的详细信息
func (this *Article) Detail(ctx *api.Context, req *dto.Article) (*api.Result[dto.Article], error) {
	data, err := service.NewArticle(ctx).GetById(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(data)
}

// Create @Summary 创建文章
// @Description 创建一篇新的文章
func (this *Article) Create(ctx *api.Context, req *dto.CreateArticle) (*api.Result[any], error) {

	_, err := service.NewArticle(ctx).Create(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// Update @Summary 更新文章
// @Description 更新已有文章的信息
func (this *Article) Update(ctx *api.Context, req *dto.UpdateArticle) (*api.Result[any], error) {
	_, err := service.NewArticle(ctx).Update(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// Delete @Summary 删除文章
// @Description 根据文章ID删除文章
func (this *Article) Delete(ctx *api.Context, req *dto.Article) (*api.Result[any], error) {
	err := service.NewArticle(ctx).Delete(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// PageList @Summary 分页查询文章列表
// @Description 分页查询文章列表，支持按分类、来源、状态、关键词等多条件筛选
func (this *Article) PageList(ctx *api.Context, req *dto.ArticleList) (*api.Result[dto.PageResult[dto.Article]], error) {
	result, err := service.NewArticle(ctx).GetPageList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}
