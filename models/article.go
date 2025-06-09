package models

import (
	"time"

	"gorm.io/gorm"
)

// Article 文章模型
type Article struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title    string `gorm:"size:255;not null;index" json:"title"` // 文章标题
	Content  string `gorm:"type:text;not null" json:"content"`    // 文章内容
	AuthorID uint   `gorm:"not null;index" json:"author_id"`      // 作者ID
	Author   User   `gorm:"foreignKey:AuthorID" json:"author"`    // 作者信息

	Likes     []Like     `json:"likes"`                                      // 点赞列表
	Favorites []Favorite `json:"favorites"`                                  // 收藏列表
	Comments  []Comment  `json:"comments"`                                   // 评论列表
	Tags      []Tag      `gorm:"many2many:article_tags" json:"tags"`         // 文章标签
	Pictures  []Picture  `gorm:"many2many:article_pictures" json:"pictures"` // 文章图片

	LikeCount     int `gorm:"default:0" json:"like_count"`     // 点赞数
	FavoriteCount int `gorm:"default:0" json:"favorite_count"` // 收藏数
	CommentCount  int `gorm:"default:0" json:"comment_count"`  // 评论数
}

// Tag 标签模型
type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name     string    `gorm:"size:50;not null;uniqueIndex" json:"name"` // 标签名称
	Articles []Article `gorm:"many2many:article_tags" json:"articles"`   // 关联的文章
}

// Like 点赞模型
type Like struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint    `gorm:"not null;index" json:"user_id"`       // 点赞用户ID
	User      User    `gorm:"foreignKey:UserID" json:"user"`       // 点赞用户
	ArticleID uint    `gorm:"not null;index" json:"article_id"`    // 被点赞文章ID
	Article   Article `gorm:"foreignKey:ArticleID" json:"article"` // 被点赞文章
}

// Favorite 收藏模型
type Favorite struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint    `gorm:"not null;index" json:"user_id"`       // 收藏用户ID
	User      User    `gorm:"foreignKey:UserID" json:"user"`       // 收藏用户
	ArticleID uint    `gorm:"not null;index" json:"article_id"`    // 被收藏文章ID
	Article   Article `gorm:"foreignKey:ArticleID" json:"article"` // 被收藏文章
}

func (Article) TableName() string {
	return "articles"
}

func (Tag) TableName() string {
	return "tags"
}

func (Like) TableName() string {
	return "likes"
}

func (Favorite) TableName() string {
	return "favorites"
}
