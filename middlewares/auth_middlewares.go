package middlewares

import (
	"net/http"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/appabin/greenbook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "缺少Authorization头"})
			ctx.Abort()
			return
		}

		username, err := utils.ParseJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":  "无效的Token",
				"detail": err.Error(), // 提供具体错误信息
			})
			ctx.Abort()
			return
		}

		// 替换原来的GetUserIDByUsername调用
		var user models.User
		if err := global.Db.
			Select("id").
			Where("username = ?", username).
			First(&user).Error; err != nil {

			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			}
			ctx.Abort()
			return
		}

		ctx.Set("username", username)
		ctx.Set("userID", user.ID) // 使用查询到的ID
	}
}
