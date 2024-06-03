package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
)

func SignUp(p *models.ParamSignUp) bool {
	//user whether or not sign up this web
	if exist := mysql.QueryUserByUsername(p.Username); exist {
		zap.L().Warn("this user did sign,Username =", zap.String("Username", p.Username), zap.String("Password", p.Password))
		return false
	}

	//make UID
	userId := snowflake.GenID()

	//save in mysql
	u := models.User{
		UserId:   userId,
		Username: p.Username,
		Password: p.Password,
	}
	mysql.InsertUser(&u)
	return true
}

func Login(p *models.ParamLogin) (user *models.User, err error) {
	user = &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 传递的是指针，就能拿到user.UserID
	flag := mysql.Login(user)
	if flag {
		return user, errors.New("username or password is error")
	}

	// 生成JWT
	token, err := jwt.GenToken(user.UserId, user.Username)
	if err != nil {
		return
	}
	user.Token = token
	return user, nil
}

func CheckParams(c *gin.Context, p *models.ParamSignUp) bool {

	err := c.ShouldBindJSON(p)

	if err != nil {
		zap.L().Error("Sign Up Handler with invalid param", zap.Error(err))
		//whether error type is validator.ValidationErrors
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"msg": err.Error(),
			})
			return false
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg": removeTopStruct(errs.Translate(trans)),
			})
		}
		return false
	}
	if len(p.RePassword) != len(p.Password) {
		zap.L().Info("Sign Up Handler with invalid param")
		c.JSON(http.StatusOK, gin.H{
			"msg": "params error",
		})
		return false
	}
	if p.Password != p.RePassword {
		zap.L().Info("Sign Up RePassword not equals Password")
		c.JSON(http.StatusOK, gin.H{
			"msg": "RePassword not equal Password",
		})
		return false
	}
	return true
}
