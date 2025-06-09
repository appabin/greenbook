package global

import (
	"context"
	"log"

	"github.com/go-redis/redis"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

// MinIOConfig MinIO 配置结构
type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

var (
	Db          *gorm.DB
	RedisDB     *redis.Client
	MinIOClient *minio.Client
	MinIOConf   *MinIOConfig
)

// InitMinIO 初始化 MinIO 客户端
func InitMinIO() {
	minioClient, err := minio.New(MinIOConf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(MinIOConf.AccessKey, MinIOConf.SecretKey, ""),
		Secure: MinIOConf.UseSSL,
	})
	if err != nil {
		log.Fatalln("MinIO 客户端初始化失败:", err)
	}
	MinIOClient = minioClient

	// 确保存储桶存在
	err = MinIOClient.MakeBucket(context.Background(), MinIOConf.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := MinIOClient.BucketExists(context.Background(), MinIOConf.BucketName)
		if errBucketExists == nil && exists {
			log.Printf("存储桶 %s 已存在\n", MinIOConf.BucketName)
		} else {
			log.Fatalln("创建存储桶失败:", err)
		}
	} else {
		log.Printf("存储桶 %s 创建成功\n", MinIOConf.BucketName)
	}
}
