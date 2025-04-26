package dao

import (
	"ai-cloud/internal/model"

	"gorm.io/gorm"
)

type UserDao interface {
	CheckFieldExists(field string, value interface{}) (bool, error)
	CreateUser(user *model.User) error
	GetUserByName(name string) (user *model.User, err error)
}

type userDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &userDao{db: db}
}

// CheckFieldExists 检查字段是否存在
func (ud *userDao) CheckFieldExists(field string, value interface{}) (bool, error) {
	var count int64
	if err := ud.db.Model(&model.User{}).Where(field+" = ?", value).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (ud *userDao) CreateUser(user *model.User) error {
	result := ud.db.Create(user)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (ud *userDao) GetUserByName(name string) (*model.User, error) {
	var user model.User
	result := ud.db.Model(&model.User{}).Where("username = ?", name).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
