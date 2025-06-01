package router

import (
	"github.com/appabin/greenbook/controllers"
	"github.com/appabin/greenbook/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 认证相关路由组
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", controllers.Login)              // 网页登录
		authGroup.POST("/register", controllers.Register)        // 网页注册
		authGroup.POST("/wechat-login", controllers.WeChatLogin) // 微信登录（新增）
	}

	// 受保护API路由组（需要JWT认证）
	apiProtected := r.Group("/api")
	apiProtected.Use(middlewares.AuthMiddleWare()) // 统一应用认证中间件
	{
		userGroup := apiProtected.Group("/user")
		{
			userGroup.GET("/info", controllers.GetCurrentUserInfo)
			userGroup.GET("/:id", controllers.GetUserProfile)
		}

		followGroup := apiProtected.Group("/follow")
		{
			followGroup.POST("", controllers.FollowAction)              // 关注/取消关注
			followGroup.GET("/following", controllers.GetFollowingList) // 关注列表
			followGroup.GET("/followers", controllers.GetFollowersList) // 粉丝列表
		}

		articleGroup := apiProtected.Group("/article")
		{
			articleGroup.POST("", controllers.CreateArticle)
			articleGroup.GET("", controllers.GetArticleList)
			articleGroup.GET("/:id", controllers.GetArticle)
		}

	}

	return r
}
