package article_test

import (
	"encoding/json"
	"fmt"
	"megin/internal/base"
	"megin/internal/config"
	"megin/internal/module/article/model"
	repo "megin/internal/module/article/repository"
	"megin/pkg/context/api"
	"testing"

	"megin/internal"
	"megin/internal/module/article/dto"
	"megin/pkg/bootstrap"
	"megin/test"
)

var (
	articleId int
)

// TestMain 测试入口函数，初始化服务
func TestMain(m *testing.M) {
	bootstrap.ServerInitWithMode("../../config/config-dev.yaml", config.RunModeMixed, internal.OnServerStart)
	m.Run()
}

// 这个事务里面，没有手动提交commit和rollback,但是可以正确执行。在defer tx.AutoCommitHandler(&err) 内部处理。可以减少很多代码
func ArticleCreateTx(tx *base.TX) (err error) {
	//自动处理事务的提交和回滚
	defer tx.AutoCommitHandler(&err)

	context := &api.Context{}
	useTx := repo.NewArticle(context).EnableTx(tx)
	artModel1 := &model.Article{
		Content:     "测试文章内容",
		OrigContent: "原始测试文章内容",
		Title:       "测试文章标题",
		OrgTitle:    "原始测试文章标题",
		Summary:     "这是一篇测试文章的摘要",
		TextType:    1,
	}

	err = useTx.Save(artModel1)
	if err != nil {
		return err
	}

	artModel2 := &model.Article{
		Content:     "测试文章内容",
		OrigContent: "原始测试文章内容",
		Title:       "测试文章标题",
		OrgTitle:    "原始测试文章标题",
		Summary:     "这是一篇测试文章的摘要",
		TextType:    1,
	}

	// TextType为tinyInt类型 。故意赋个超出界限的值，让他出错。测试回滚
	//artModel2.TextType = 11111111

	err = useTx.Save(artModel2)
	return err
}

func TestArticleCreateTx(t *testing.T) {
	tx := base.TxBegin()
	err := ArticleCreateTx(tx)

	fmt.Println(err)
}

func TestArticleCreate(t *testing.T) {
	req := dto.CreateArticle{
		Content:     "测试文章内容",
		OrigContent: "原始测试文章内容",
		Title:       "测试文章标题",
		OrgTitle:    "原始测试文章标题",
		Summary:     "这是一篇测试文章的摘要",
		Category1:   1, // 替换为实际存在的分类ID
		Category2:   2, // 替换为实际存在的分类ID
		Source:      1, // 替换为实际存在的来源ID
		OrigLink:    "https://example.com/test-article",
	}
	resp := test.Post("/api/article/create", req)

	// 解析响应，获取创建的文章ID
	var result struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    dto.Article `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err == nil && result.Success {
		articleId = result.Data.ID
		t.Logf("成功创建文章，ID: %d", articleId)
	}

	test.Print(resp.Body.String())
}

func TestArticleUpdate(t *testing.T) {

	req := dto.UpdateArticle{
		ID:        1,
		Content:   "更新后的文章内容",
		Title:     "更新后的文章标题",
		Summary:   "更新后的文章摘要",
		Category1: 1, // 替换为实际存在的分类ID
		Category2: 2, // 替换为实际存在的分类ID
		Source:    1, // 替换为实际存在的来源ID
		OrigLink:  "https://example.com/updated-test-article",
	}
	resp := test.Post("/api/article/update", req)
	test.Print(resp.Body.String())
}

func TestArticleDetail(t *testing.T) {

	req := map[string]interface{}{
		"id": 1,
	}
	resp := test.Get("/api/article/detail", req)
	test.Print(resp.Body.String())

	resp = test.Get("/api/article/detail", req)
	test.Print(resp.Body.String())
}

func TestArticlePageList(t *testing.T) {
	req := map[string]interface{}{
		"category1": 1,  // 替换为实际存在的分类ID
		"category2": 2,  // 替换为实际存在的分类ID
		"source":    1,  // 替换为实际存在的来源ID
		"page":      1,  // 页码
		"page_size": 10, // 每页数量，注意这里使用下划线格式
	}
	resp := test.Get("/admin-api/article/pageList", req)
	test.Print(resp.Body.String())
}

func TestArticleDelete(t *testing.T) {
	if articleId == 0 {
		t.Skip("没有创建文章，跳过删除测试")
	}

	req := map[string]interface{}{
		"id": articleId,
	}
	resp := test.Post("/admin-api/article/delete", req)
	test.Print(resp.Body.String())
}
