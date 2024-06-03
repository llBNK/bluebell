package controller

import (
	"bluebell/dao/redis"
	"bluebell/logic"
	"bluebell/models"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

const Login_Token = "login_token"

type TokenColl struct {
	FirstToken  string
	SecondToken string
	ThirdToken  string
}

func SignUpHandler(c *gin.Context) {
	//1. params check
	p := new(models.ParamSignUp)
	if !logic.CheckParams(c, p) {
		return
	}
	//2.business handler
	if flag := logic.SignUp(p); !flag {
		ResponseErrorWithMsg(c, CodeUserExist, CodeUserExist.Msg())
		return
	}
	//3.return response
	c.JSON(http.StatusOK, "ok")
}

// LoginHandler 登录
func LoginHandler(c *gin.Context) {
	// 1.获取请求参数及参数校验
	p := new(models.ParamLogin)
	p.Username = c.Query("username")
	p.Password = c.Query("password")

	//if err := c.ShouldBindJSON(p); err != nil {
	//	// 请求参数有误，直接返回响应
	//	zap.L().Error("Login with invalid param", zap.String("username",p.Username),zap.Error(err))
	//	// 判断err是不是validator.ValidationErrors 类型
	//	_, ok := err.(validator.ValidationErrors)
	//	if !ok {
	//		ResponseError(c, CodeInvalidParam)
	//		return
	//	}
	//	ResponseErrorWithMsg(c, CodeInvalidParam, "请求参数错误")
	//	return
	//}
	// 2.业务逻辑处理
	user, err := logic.Login(p)
	if err != nil {
		zap.L().Error("logic.Login failed", zap.String("username", p.Username), zap.Error(err))
		ResponseError(c, CodeInvalidPassword)
		return
	}

	//3.在返回前去将token放入redis中
	flag := setRedisTokenV2(user)
	if !flag {
		ResponseError(c, CodeInvalidPassword)
		return
	}
	// 4.返回响应
	ResponseSuccess(c, gin.H{
		"user_id":   fmt.Sprintf("%d", user.UserId), // id值大于1<<53-1  int64类型的最大值是1<<63-1
		"user_name": user.Username,
		"token":     user.Token,
	})
}

func setRedisTokenV2(user *models.User) bool {
	res, err := redis.Rdb.Get(fmt.Sprintf(Login_Token+"_"+"%s", user.Username)).Result()
	if err != nil && err.Error() != redis.NIl {
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
		if tokenstr == user.Token {
			return true
		}
	}
	if len(tokens) < 3 {
		//tokens[len(tokens)] = user.Token
		tokens = append(tokens, user.Token)
	} else {
		copy(tokens[1:], tokens[:2])
		tokens[0] = user.Token
	}

	// 将 slice 转换为 JSON 字符串
	jsonData, err := json.Marshal(tokens)
	if err != nil {
		log.Fatalf("Error marshalling to JSON: %v", err)
	}
	jsonString := string(jsonData)
	fmt.Println("JSON string:", jsonString)
	res, err = redis.Rdb.Set(fmt.Sprintf(Login_Token+"_%s", user.Username), jsonString, 3600*time.Second).Result()

	if err != nil {
		return false
	}
	return true
}

func setRedisToken(user *models.User) bool {
	res, err := redis.Rdb.Get(fmt.Sprintf(Login_Token+"_"+"%s", user.Username)).Result()
	if err != nil {
		if err.Error() != redis.NIl {
			return false
		}
	}
	tokenColl := new(TokenColl)

	if res != "" {
		err = json.Unmarshal([]byte(res), &tokenColl)
		if err != nil {
			return false
		}
	}

	if user.Token == tokenColl.FirstToken || user.Token == tokenColl.SecondToken || user.Token == tokenColl.ThirdToken {
		return true
	}

	if tokenColl.FirstToken == "" {
		tokenColl.FirstToken = user.Token

	} else if tokenColl.SecondToken == "" {
		tokenColl.SecondToken = user.Token

	} else if tokenColl.ThirdToken == "" {
		tokenColl.ThirdToken = user.Token

	} else {
		setUserTokens(tokenColl, user.Token)
	}

	marshal, err := json.Marshal(tokenColl)
	if err != nil {
		return false
	}
	res, err = redis.Rdb.Set(fmt.Sprintf(Login_Token+"_%s", user.Username), marshal, 3600*time.Second).Result()

	if err != nil {
		return false
	}
	return true
}

func setUserTokens(tokens *TokenColl, token string) {
	tokens.ThirdToken = tokens.SecondToken
	tokens.SecondToken = tokens.FirstToken
	tokens.SecondToken = token
}
