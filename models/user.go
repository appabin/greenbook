package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型，包含微信小程序相关字段
// @Description 用户信息
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`              // 用户ID
	CreatedAt time.Time      `json:"created_at"`                        // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                        // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // 删除时间，使用 gorm.DeletedAt 类型

	Username string `gorm:"size:50;comment:用户名" json:"username"`
	Password string `gorm:"size:100;comment:密码" json:"password"`
	
	Nickname string `gorm:"size:50;comment:昵称" json:"nickname"`
	Avatar   string `gorm:"size:255;comment:头像URL" json:"avatar"`
	Gender   uint8  `gorm:"default:0;comment:性别(0:未知,1:男,2:女)" json:"gender"`
	Phone    string `gorm:"size:20;comment:手机号" json:"phone"`
	Email    string `gorm:"size:100;comment:邮箱" json:"email"`

	// 微信小程序相关字段
	OpenID     string `gorm:"size:50;uniqueIndex;comment:微信openid" json:"open_id"`
	UnionID    string `gorm:"size:50;comment:微信unionid" json:"union_id"`
	SessionKey string `gorm:"size:100;comment:微信session_key" json:"session_key"`
	Country    string `gorm:"size:50;comment:国家" json:"country"`
	Province   string `gorm:"size:50;comment:省份" json:"province"`
	City       string `gorm:"size:50;comment:城市" json:"city"`
	Language   string `gorm:"size:20;comment:语言" json:"language"`

	// 论坛社交相关字段
	FollowersCount uint `gorm:"default:0;comment:粉丝数" json:"followers_count"`
	FollowingCount uint `gorm:"default:0;comment:关注数" json:"following_count"`
	PostsCount     uint `gorm:"default:0;comment:发帖数" json:"posts_count"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}
