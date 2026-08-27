package system

import (
	"net/http/httptest"
	"testing"

	systemDto "megin/internal/system/dto"
	"megin/pkg/context/api"

	"github.com/gin-gonic/gin"
)

func TestBlacklistTokenFallsBackToHeaderToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = httptest.NewRequest("POST", "/admin-api/jwt/jsonInBlacklist", nil)
	ginCtx.Request.Header.Set("x-token", "header-token")

	ctx, err := api.NewContext(ginCtx)
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	token := blacklistToken(ctx, &systemDto.JsonInBlacklistReq{})
	if token != "header-token" {
		t.Fatalf("blacklistToken() = %q, want %q", token, "header-token")
	}
}
