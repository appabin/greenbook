package main

import (
	"fmt"
	"log"

	"github.com/appabin/greenbook/config"
	"github.com/appabin/greenbook/global"
	"github.com/appabin/greenbook/router"
)

func main() {
	config.InitConfig()

	log.Println("=== 配置加载成功 ===")
	fmt.Printf("应用名称: %s\n", config.AppConfig.App.Name)
	fmt.Printf("应用端口: %s\n", config.AppConfig.App.Port)

	// 设置 MinIO 配置
	global.MinIOConf = &global.MinIOConfig{
		Endpoint:   config.AppConfig.MinIO.Endpoint,
		AccessKey:  config.AppConfig.MinIO.AccessKey,
		SecretKey:  config.AppConfig.MinIO.SecretKey,
		BucketName: config.AppConfig.MinIO.BucketName,
		UseSSL:     config.AppConfig.MinIO.UseSSL,
	}

	// 初始化 MinIO
	global.InitMinIO()
	log.Println("=== MinIO 初始化成功 ===")

	r := router.SetupRouter()
	
	r.Run(":" + config.AppConfig.App.Port)
}
