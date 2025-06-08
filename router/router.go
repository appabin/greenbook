package router

import (
	"github.com/appabin/greenbook/controllers"
	"github.com/appabin/greenbook/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")

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
			userGroup.POST("/avatar", controllers.UpdateUserAvatar) // 更新用户头像
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

		commentGroup := apiProtected.Group("/comment")
		{
			commentGroup.POST("/:article_id", controllers.CreateComment)
		}

		likeGroup := apiProtected.Group("/like")
		{
			likeGroup.POST("/:article_id", controllers.ArticleToggleLike) // 点赞/取消点赞
			likeGroup.POST("/comment/:comment_id", controllers.CommentToggleLike)
		}

		favoriteGroup := apiProtected.Group("/favorite")
		{
			favoriteGroup.POST("/:article_id", controllers.ArticleToggleFavorite) // 收藏/取消收藏
		}

		photoGroup := apiProtected.Group("/picture")
		{
			photoGroup.POST("/upload", controllers.UploadPicture)
			photoGroup.POST("/upload/multipart", controllers.UploadPictureMultipart)
		}
	}
	return r
}
