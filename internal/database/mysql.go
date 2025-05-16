package database

import (
	"ai-cloud/config"
	"ai-cloud/internal/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
	dbErr  error
)

// GetDB 获取数据库单例
func GetDB() (*gorm.DB, error) {
	dbOnce.Do(func() {
		// 构造 DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.AppConfigInstance.Database.User,
			config.AppConfigInstance.Database.Password,
			config.AppConfigInstance.Database.Host,
			config.AppConfigInstance.Database.Port,
			config.AppConfigInstance.Database.Name,
		)

		// 连接数据库
		db, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if dbErr != nil {
			return
		}

		// 自动迁移, 创建表结构
		if err := db.AutoMigrate(
			&model.User{},
			&model.File{},
			&model.KnowledgeBase{},
			&model.Document{},
			&model.Model{},
			&model.Agent{},
			// 会话记录相关
			&model.Conversation{},
			&model.Message{},
			&model.Attachment{},
			&model.MessageAttachment{},
		); err != nil {
			dbErr = err
			return
		}
	})

	return db, dbErr
}
