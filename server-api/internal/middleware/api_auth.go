package middleware

import (
	"errors"
	bizcache "megin/internal/cache"
	commonDto "megin/internal/dto"
	"megin/pkg/context/api"
	"megin/pkg/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApiAuthTokenRequired() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := getToken(context)

		if tokenString == "" {
			result := api.Failed[any](errors.New("token不能为空"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}
		claims := &commonDto.Claims{}
		tokenInfo, err := parseClaims(tokenString, claims)

		if tokenInfo == nil || err != nil {
			result := api.Failed[error](errs.NewBusinessError(403, "token已失效,请重新登录"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		if !tokenInfo.Valid {
			result := api.Failed[error](errs.NewBusinessError(403, "token已失效,请重新登录"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		if err != nil {
			result := api.Failed[error](errs.NewBusinessError(403, "token验证失败,请重新登录"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		redisToken, err := bizcache.GetRedisString(bizcache.GetApiUserLoginTokenKey(uint(claims.UserID)))
		if err != nil {
			result := api.Failed[error](errs.NewNormalError(500, "读取前台用户登录token失败", err))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}
		if redisToken == "" {
			result := api.Failed[error](errs.NewBusinessError(403, "登录状态已失效,请重新登录"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}
		if redisToken != tokenString {
			result := api.Failed[error](errs.NewBusinessError(403, "您的帐户已在其他设备登录,请重新登录"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		context.Set(commonDto.ApiClaimToken, tokenString)
		context.Set(commonDto.ApiJwtClaims, claims)
		context.Next()
	}
}
