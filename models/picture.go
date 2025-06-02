package models

import (
	"time"

	"gorm.io/gorm"
)

// Picture 图片模型
type Picture struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	URL       string         `gorm:"size:500;not null" json:"url"`  // 图片存储URL
	UserID    uint           `gorm:"not null;index" json:"user_id"` // 上传用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user"` // 上传用户
}

// ArticlePicture 文章图片关联表
type ArticlePicture struct {
	ArticleID uint    `gorm:"primaryKey;not null;index" json:"article_id"` // 文章ID
	PictureID uint    `gorm:"primaryKey;not null;index" json:"picture_id"` // 图片ID
	Order     int     `gorm:"not null;default:0" json:"order"`             // 图片在文章中的顺序，0为封面
	Article   Article `gorm:"foreignKey:ArticleID" json:"article"`
	Picture   Picture `gorm:"foreignKey:PictureID" json:"picture"`
}

func (Picture) TableName() string {
	return "pictures"
}

func (ArticlePicture) TableName() string {
	return "article_pictures"
}
