package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"megin/pkg/context/api"

	"github.com/gin-gonic/gin"
)

type queryPage struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type queryRequest struct {
	queryPage
	IDs  []uint    `form:"ids"`
	Name *string   `form:"name"`
	Date time.Time `form:"date" time_format:"2006-01-02"`
	ID   uint      `path:"id" binding:"required"`
}

func TestBindParamsSupportsEmbeddedSlicePointerTimeAndPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/items/7?page=2&pageSize=20&ids=1&ids=3&name=test&date=2026-06-24", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "7"}}

	var got queryRequest
	if err := BindParams(c, &got); err != nil {
		t.Fatalf("BindParams() error = %v", err)
	}
	if got.Page != 2 || got.PageSize != 20 || got.ID != 7 {
		t.Fatalf("unexpected scalar values: %+v", got)
	}
	if len(got.IDs) != 2 || got.IDs[0] != 1 || got.IDs[1] != 3 {
		t.Fatalf("unexpected IDs: %#v", got.IDs)
	}
	if got.Name == nil || *got.Name != "test" {
		t.Fatalf("unexpected Name: %#v", got.Name)
	}
	if got.Date.Format("2006-01-02") != "2026-06-24" {
		t.Fatalf("unexpected Date: %v", got.Date)
	}
}

func TestWrapHandlerForParamsWritesOneResponseOnError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := WrapHandlerForParams(func(_ *api.Context, _ struct{}) (map[string]any, error) {
		return nil, errors.New("failed")
	})
	handler(c)

	if !json.Valid(w.Body.Bytes()) {
		t.Fatalf("handler wrote multiple or invalid JSON responses: %q", w.Body.String())
	}
}

func TestWrapHandlerWithBodyRejectsEmptyPointerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(nil))
	c.Request.Header.Set("Content-Type", "application/json")

	type reqBody struct {
		Token string `json:"token" binding:"required"`
	}

	called := false
	handler := WrapHandlerWithBody(func(_ *api.Context, req *reqBody) (map[string]any, error) {
		called = true
		if req == nil {
			t.Fatal("request should be initialized before handler")
		}
		return map[string]any{"token": req.Token}, nil
	})
	handler(c)

	if called {
		t.Fatal("handler should not be called when required body is empty")
	}
	if !json.Valid(w.Body.Bytes()) {
		t.Fatalf("handler wrote invalid JSON response: %q", w.Body.String())
	}
}
