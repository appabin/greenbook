package models

import (
	"time"

	"gorm.io/gorm"
)

// UserFollow 用户关系模型（关注/粉丝）
type UserFollow struct {
	FollowerID uint           `gorm:"primaryKey;index" json:"follower_id"` // 关注者ID
	FollowedID uint           `gorm:"primaryKey;index" json:"followed_id"` // 被关注者ID
	CreatedAt  time.Time      `json:"created_at"`                          // 创建时间
	UpdatedAt  time.Time      `json:"updated_at"`                          // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                       // 软删除时间
}

// TableName 设置表名
func (UserFollow) TableName() string {
	return "user_follows"
}
