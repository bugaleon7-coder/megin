package middleware

import (
	"errors"
	commonDto "megin/internal/module/common/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
	"megin/pkg/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminApiAuthTokenRequired() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := getToken(context)

		if tokenString == "" {
			result := api.Failed[any](errors.New("token不能为空"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		blacklisted, err := systemService.IsBlacklistedJWT(tokenString)
		if err != nil {
			result := api.Failed[error](errs.NewBusinessError(403, "token验证失败,请重新登录"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}
		if blacklisted {
			result := api.Failed[error](errs.NewBusinessError(403, "您的帐户异地登陆或令牌失效"))
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

		context.Set(commonDto.AdminApiClaimToken, tokenString)
		context.Set(commonDto.AdminApiJwtClaims, claims)

		allowed, err := systemService.EnforceAuthorityPolicy(uint(claims.RoleId), context.Request.URL.Path, context.Request.Method)
		if err != nil {
			result := api.Failed[error](errs.NewBusinessError(403, "权限验证失败"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}
		if !allowed {
			if systemService.IsDefaultAuthenticatedPolicy(context.Request.URL.Path, context.Request.Method) {
				context.Next()
				return
			}
			result := api.Failed[error](errs.NewBusinessError(403, "权限不足"))
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		context.Next()
	}
}
