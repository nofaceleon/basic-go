package middleware

import (
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

// LoginMiddlewareBuilder 扩展性
type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}
func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		// 不需要登录校验的
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}
		//sess := sessions.Default(ctx)
		//id := sess.Get("userId")
		//这边改成使用jwt, 获取jwt, 提取userid

		authorization := ctx.GetHeader("Authorization")

		if authorization == "" {
			//说明没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//提取token
		list := strings.Split(authorization, " ") // 通过空格进行切分
		if len(list) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := list[1]
		userClaims := &web.UserClaims{} //这边传递的是一个指针
		//解析token
		token, err := jwt.ParseWithClaims(tokenStr, userClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte("AllYourBase"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid || userClaims.UserId == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("claims", userClaims) //将解析出来的用户信息保存到上下文中
	}
}

var IgnorePaths []string

func CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range IgnorePaths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 不需要登录校验的
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//已经登录, 如何

	}
}

func CheckLoginV1(paths []string, abc int, bac int64, asdsd string) gin.HandlerFunc {
	if len(paths) == 0 {
		paths = []string{}
	}
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 不需要登录校验的
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
