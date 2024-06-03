package routes

import (
	"bluebell/controller"
	"bluebell/logger"
	"bluebell/logic"
	"bluebell/middlewares"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // gin设置成发布模式
	}
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinLogger())

	if err := logic.InitTrans("zh"); err != nil {
		zap.L().Error("init translation error", zap.Error(err))
	}

	//login route
	r.POST("/signup", controller.SignUpHandler)

	// 登录
	r.GET("/login", controller.LoginHandler)

	r.GET("/ping", middlewares.JWTAuthMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, "ok")
	})

	return r

}
