package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string `mapstructure:"name"`
		Port string `mapstructure:"port"`
	} `mapstructure:"app"`
	
	Wechat struct {
		AppID     string `mapstructure:"app_id"`
		AppSecret string `mapstructure:"app_secret"`
	} `mapstructure:"wechat"`
	
	Database struct {
		Dsn           string `mapstructure:"dsn"`
		MaxIdleConns  int
		MaxIdOpenCons int
	} `mapstructure:"database"`
	Redis struct {
		Addr     string
		DB       int
		Password string
	} `mapstructure:"redis"`
	MinIO struct {
		Endpoint   string `mapstructure:"endpoint"`
		AccessKey  string `mapstructure:"access_key"`
		SecretKey  string `mapstructure:"secret_key"`
		BucketName string `mapstructure:"bucket_name"`
		UseSSL     bool   `mapstructure:"use_ssl"`
	} `mapstructure:"minio"`
}

var AppConfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml") // 推荐使用 yaml 而不是 yml
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %v", err))
	}

	// 绑定到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic(fmt.Errorf("配置解析失败: %v", err))
	}

	// 打印加载成功的配置
	log.Println("配置文件加载成功")
	initDB()
	InitRedis()
}

// 在包外使用时添加打印示例
