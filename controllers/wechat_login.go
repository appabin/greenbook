package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/appabin/greenbook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WeChatLoginRequest 微信登录请求结构
type WeChatLoginRequest struct {
	Code          string `json:"code" binding:"required"` // 微信登录code
	EncryptedData string `json:"encryptedData"`           // 加密数据（首次登录需要）
	IV            string `json:"iv"`                      // 解密向量（首次登录需要）
	Phone         string `json:"phone"`                   // 明文手机号（备用）
}

// WeChatLogin 微信用户登录/注册
func WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 获取微信配置（建议从配置读取）
	appID := "YOUR_APP_ID"
	appSecret := "YOUR_APP_SECRET"

	// 获取微信会话信息
	sessionRes, err := utils.GetWeChatSession(appID, appSecret, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "微信登录失败", "detail": err.Error()})
		return
	}
	//可能有问题 
	// 查询已有用户
	var user models.User
	dbResult := global.Db.Where("open_id = ?", sessionRes.OpenID).First(&user)

	// 处理首次登录用户
	if dbResult.Error != nil && errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		// 解密手机号（首次登录需要）
		var phone string
		if req.EncryptedData != "" && req.IV != "" {
			decrypted, err := utils.DecryptWeChatData(sessionRes.SessionKey, req.EncryptedData, req.IV)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "数据解密失败"})
				return
			}
			phone = fmt.Sprintf("%v", decrypted["purePhoneNumber"])
		} else if req.Phone != "" {
			phone = req.Phone
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "需要绑定手机号"})
			return
		}

		// 创建新用户
		user = models.User{
			OpenID:     &sessionRes.OpenID,
			UnionID:    sessionRes.UnionID,
			SessionKey: sessionRes.SessionKey,
			Phone:      phone,
			// 设置默认值
			Nickname: "微信用户",
			Avatar:   "https://default.avatar",
		}

		if err := global.Db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
			return
		}
	} else if dbResult.Error == nil {
		// 更新会话信息
		global.Db.Model(&user).Updates(map[string]interface{}{
			"session_key": sessionRes.SessionKey,
		})
	}

	// 生成JWT
	token, err := utils.GenerateJWT(fmt.Sprintf("%d", user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 返回结果（过滤敏感字段）
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"phone":    user.Phone,
		},
	})
}
