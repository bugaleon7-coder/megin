package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFiles_PreserveMethodCommentsByReceiver(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	file := filepath.Join(dir, "handler.go")
	content := `package sample

type User struct{}
type SysUser struct{}

// @Summary 前台用户登录
// @Description 前台登录描述
func (h *User) Login() {}

// @Summary 用户登录
// @Description 用户登录，返回JWT Token
func (h *SysUser) Login() {}
`
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	comments, err := ParseFiles(file)
	if err != nil {
		t.Fatalf("解析测试文件失败: %v", err)
	}

	if got := comments.FuncSummary["User.Login"]; got != "前台用户登录" {
		t.Fatalf("User.Login 摘要错误，got=%q", got)
	}
	if got := comments.FuncDescription["User.Login"]; got != "前台登录描述" {
		t.Fatalf("User.Login 描述错误，got=%q", got)
	}
	if got := comments.FuncSummary["SysUser.Login"]; got != "用户登录" {
		t.Fatalf("SysUser.Login 摘要错误，got=%q", got)
	}
	if got := comments.FuncDescription["SysUser.Login"]; got != "用户登录，返回JWT Token" {
		t.Fatalf("SysUser.Login 描述错误，got=%q", got)
	}
}
