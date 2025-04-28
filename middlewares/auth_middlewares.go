package middlewares

import (
	"errors"
	"net/http"
	"strconv"

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

		userIDStr, err := utils.ParseJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":  "无效的Token",
				"detail": err.Error(),
			})
			ctx.Abort()
			return
		}

		// 转换userID为uint
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID格式"})
			ctx.Abort()
			return
		}

		// 查询用户是否存在
		var user models.User
		if err := global.Db.
			Select("id, username"). // 获取必要字段
			Where("id = ?", uint(userID)).
			First(&user).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			}
			ctx.Abort()
			return
		}

		// 设置上下文信息
		ctx.Set("username", user.Username)
		ctx.Set("userID", user.ID)
		ctx.Next()
	}
}
