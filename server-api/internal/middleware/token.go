package middleware

import (
	"megin/internal/config"
	commonDto "megin/internal/module/common/dto"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getToken(ctx *gin.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return authHeader[7:]
	}

	for _, header := range []string{"X-Token", "x-token", "Token"} {
		if token := ctx.GetHeader(header); token != "" {
			return token
		}
	}

	token, _ := ctx.Cookie("x-token")
	return token
}

func parseClaims(tokenString string, claims *commonDto.Claims) (*jwt.Token, error) {
	conf := config.GetConfig()
	secrets := []string{conf.Jwt.Secret, "gva-jwt-secret"}
	var lastErr error

	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		currentClaims := &commonDto.Claims{}
		tokenInfo, err := jwt.ParseWithClaims(tokenString, currentClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err == nil && tokenInfo != nil && tokenInfo.Valid {
			*claims = *currentClaims
			return tokenInfo, nil
		}
		lastErr = err
	}

	return nil, lastErr
}
