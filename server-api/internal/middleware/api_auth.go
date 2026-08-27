package middleware

import (
	"errors"
	commonDto "megin/internal/module/common/dto"
	userService "megin/internal/module/user/service"
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

		requestCtx, ctxErr := api.NewContext(context)
		if ctxErr != nil {
			result := api.Failed[error](ctxErr)
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}
		if err := userService.NewUser(requestCtx).ValidateLoginToken(uint(claims.UserID), tokenString); err != nil {
			result := api.Failed[error](err)
			context.JSON(http.StatusOK, result)
			context.Abort()
			return
		}

		context.Set(commonDto.ApiClaimToken, tokenString)
		context.Set(commonDto.ApiJwtClaims, claims)
		context.Next()
	}
}
