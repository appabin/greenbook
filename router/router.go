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

	api := r.Group("api")

	// 文章相关路由
	articleController := &controllers.ArticleController{}
	articles := api.Group("/articles")
	{
		// 无需登录的路由
		articles.GET("", articleController.List)    // 获取文章列表
		articles.GET("/:id", articleController.Get) // 获取文章详情

		// 需要登录的路由
		auth := articles.Use(middlewares.AuthMiddleWare())
		{
			auth.POST("", articleController.Create)                     // 创建文章
			auth.PUT("/:id", articleController.Update)                  // 更新文章
			auth.DELETE("/:id", articleController.Delete)               // 删除文章
			auth.POST("/:id/like", articleController.ToggleLike)        // 点赞/取消点赞
			auth.POST("/:id/comments", articleController.CreateComment) // 发表评论
		}
	}

	api.Use(middlewares.AuthMiddleWare())
	{
		// api.POST("/follow/:id", controllers.FollowUser)     // 关注用户
		// api.DELETE("/follow/:id", controllers.UnfollowUser) // 取消关注
		// api.GET("/followings", controllers.GetFollowings)   // 获取关注列表
		// api.GET("/followers", controllers.GetFollowers)     // 获取粉丝列表
	}

	return r
}
