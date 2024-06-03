package middlewares

import (
	"bluebell/controller"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// Authorization: Bearer xxxxxxx.xxx.xxx  / X-TOKEN: xxx.xxx.xx
		// 这里的具体实现方式要依据你的实际业务情况决定
		authHeader := c.Request.Header.Get("Authorization")
		p := new(models.ParamLogin)
		err := c.ShouldBindJSON(p)
		if err != nil {
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort()
			return
		}
		if authHeader == "" {
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort()
			return
		}

		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		//一个账号可以三台登录
		flag := accountNumberLimitV2(parts[1], p)
		if !flag {
			controller.ResponseError(c, controller.AccountNumberOverLimit)
			c.Abort()
			return
		}
		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set(controller.CtxUserIDKey, mc.UserID)

		c.Next() // 后续的处理请求的函数中 可以用过c.Get(CtxUserIDKey) 来获取当前请求的用户信息
	}
}

func accountNumberLimitV2(authHeader string, p *models.ParamLogin) bool {
	res, err := redis.Rdb.Get(fmt.Sprintf(controller.Login_Token+"_"+"%s", p.Username)).Result()
	if err != nil {
		return false
	}
	var tokens []string
	if res != "" {
		err = json.Unmarshal([]byte(res), &tokens)
		if err != nil {
			return false
		}
	}
	fmt.Println("Slice from JSON:", tokens)
	for _, tokenstr := range tokens {
		if tokenstr == authHeader {
			return true
		}
	}
	return false
}

func accountNumberLimit(authHeader string, p *models.ParamLogin) bool {
	res, err := redis.Rdb.Get(fmt.Sprintf(controller.Login_Token+"_"+"%s", p.Username)).Result()
	if err != nil {
		return false
	}
	var tokens controller.TokenColl
	err = json.Unmarshal([]byte(res), &tokens)
	if err != nil {
		return false // 反序列化失败也应返回错误
	}
	if tokens.ThirdToken != authHeader && tokens.SecondToken != authHeader && tokens.FirstToken != authHeader {
		return false
	}
	return true
}
