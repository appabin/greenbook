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

	// // 公共API路由组（无需认证）
	// apiPublic := r.Group("/api")
	// {
	// 	// 文章公开接口
	// 	articlePublic := apiPublic.Group("/articles")
	// 	{
	// 		articlePublic.GET("", controllers.ArticleController.List)    // 文章列表
	// 		articlePublic.GET("/:id", controllers.ArticleController.Get) // 文章详情
	// 	}
	// }

	// // 受保护API路由组（需要JWT认证）
	apiProtected := r.Group("/api")
	apiProtected.Use(middlewares.AuthMiddleWare()) // 统一应用认证中间件
	{
		// 	// 文章操作接口
		// 	articleProtected := apiProtected.Group("/articles")
		// 	{
		// 		articleProtected.POST("", controllers.ArticleController.Create)              // 创建文章
		// 		articleProtected.PUT("/:id", controllers.ArticleController.Update)           // 更新文章
		// 		articleProtected.DELETE("/:id", controllers.ArticleController.Delete)        // 删除文章
		// 		articleProtected.POST("/:id/like", controllers.ArticleController.ToggleLike) // 点赞操作

		// 		// 评论相关子组
		// 		commentsGroup := articleProtected.Group("/:id/comments")
		// 		{
		// 			commentsGroup.POST("", controllers.ArticleController.CreateComment)               // 发表评论
		// 			commentsGroup.DELETE("/:comment_id", controllers.ArticleController.DeleteComment) // 删除评论
		// 		}
		// 	}

		// 	// 用户社交功能（保持注释状态，待实现）
		// 	// apiProtected.POST("/follow/:id", controllers.FollowUser)
		// 	// apiProtected.DELETE("/follow/:id", controllers.UnfollowUser)
		// 	// apiProtected.GET("/followings", controllers.GetFollowings)
		// 	// apiProtected.GET("/followers", controllers.GetFollowers)
		followGroup := apiProtected.Group("/follow")
		{
			followGroup.POST("", controllers.FollowAction)              // 关注/取消关注
			followGroup.GET("/following", controllers.GetFollowingList) // 关注列表
			followGroup.GET("/followers", controllers.GetFollowersList) // 粉丝列表
		}
	}
	return r
}
