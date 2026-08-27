package convert

import (
	articleDto "megin/internal/module/article/dto"
	articleModel "megin/internal/module/article/model"
)

// ToArticleDTO 将文章模型转换为文章 DTO，统一收口领域层对外结构转换逻辑。
func ToArticleDTO(article articleModel.Article) articleDto.Article {
	return articleDto.Article{
		ID:        article.ID,
		TextType:  article.TextType,
		Content:   article.Content,
		Title:     article.Title,
		Summary:   article.Summary,
		Category1: article.Category1,
		Category2: article.Category2,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
		Source:    article.Source,
		OrigLink:  article.OrigLink,
	}
}
