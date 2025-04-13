package router

import (
	"github.com/appabin/greenbook/controllers"
	"github.com/appabin/greenbook/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	auth := r.Group("api/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
	}

	// 微信小程序相关路由
	wx := r.Group("api/wx")
	{
		wx.POST("/login", controllers.WechatLogin)
	}

	api := r.Group("/api")

	// api.GET("/articles/:id/comments", controllers.GetArticleComments)

	api.Use(middlewares.AuthMiddleWare())
	{
		// api.POST("/articles", controllers.CreateArticle)
		// api.GET("/articles", controllers.GetArticles)
		// api.GET("/articles/:id", controllers.GetArticleByID)

		// api.POST("/articles/:id/like", controllers.LikeArticle)
		// api.GET("/articles/:id/like", controllers.GetArticleLikes)

		// api.POST("/comments", controllers.CreateComment)
		// api.DELETE("/comments/:id", controllers.DeleteComment)

		// // 新增关注相关路由
		// api.POST("/follow/:id", controllers.FollowUser)     // 关注用户
		// api.DELETE("/follow/:id", controllers.UnfollowUser) // 取消关注
		// api.GET("/followings", controllers.GetFollowings)   // 获取关注列表
		// api.GET("/followers", controllers.GetFollowers)     // 获取粉丝列表
	}

	return r
}
