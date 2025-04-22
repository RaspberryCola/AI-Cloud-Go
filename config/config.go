package config

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var AppConfigInstance *AppConfig

// InitConfig 初始化配置
func InitConfig() {
	// 加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// 监听配置变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(&AppConfigInstance); err != nil {
			log.Printf("loadConfig failed, unmarshal config err: %v", err)
		}
	})

	// 解析配置
	if err := viper.Unmarshal(AppConfigInstance); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
}

// GetConfig 获取配置
func GetConfig() *AppConfig {
	return AppConfigInstance
}
