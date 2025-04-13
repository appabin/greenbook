package controllers

import (
	"net/http"
	"time"

	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/models"
	"github.com/appabin/greenbook/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// WechatLoginRequest 微信小程序登录请求参数
type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	Gender    uint8  `json:"gender"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Language  string `json:"language"`
}

// WechatCode2SessionResponse 微信登录凭证校验返回结果
type WechatCode2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WechatLogin 微信小程序登录
func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查询用户是否存在
	var user models.User
	if err := global.Db.Where("open_id = ?", req.Code).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，创建新用户
			user = models.User{
				OpenID:   req.Code,
				Nickname: req.Nickname,
				Avatar:   req.AvatarURL,
				Gender:   req.Gender,
				Country:  req.Country,
				Province: req.Province,
				City:     req.City,
				Language: req.Language,
				Username: req.Code, // 使用openid作为用户名
			}

			if err := global.Db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			return
		}
	}

	// 更新用户信息
	user.Nickname = req.Nickname
	user.Avatar = req.AvatarURL
	user.Gender = req.Gender
	user.Country = req.Country
	user.Province = req.Province
	user.City = req.City
	user.Language = req.Language
	user.LastLoginAt = time.Now()

	if err := global.Db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查询用户
	var user models.User
	if err := global.Db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		}
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// Register 用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := global.Db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HassPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建新用户
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	if dbErr := global.Db.Create(&user).Error; dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}
