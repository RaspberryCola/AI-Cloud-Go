package dao

import (
	"ai-cloud/internal/model"
	"context"
	"errors"
	"gorm.io/gorm"
)

type ModelDao interface {
	Create(ctx context.Context, m *model.Model) error
	Update(ctx context.Context, m *model.Model) error
	Delete(ctx context.Context, userID uint, modelID string) error
	GetByID(ctx context.Context, userID uint, modelID string) (*model.Model, error)
	List(ctx context.Context, userID uint, modelType string) ([]*model.Model, error)
	Page(ctx context.Context, userID uint, modelType string, page, size int) ([]*model.Model, int64, error)
}

type modelDao struct {
	db *gorm.DB
}

func NewModelDao(db *gorm.DB) ModelDao {
	return &modelDao{db: db}
}

func (d *modelDao) Create(ctx context.Context, m *model.Model) error {
	return d.db.WithContext(ctx).Create(m).Error
}

func (d *modelDao) Update(ctx context.Context, m *model.Model) error {
	// 检查模型是否属于该用户
	var count int64
	if err := d.db.WithContext(ctx).Model(&model.Model{}).Where("id = ? AND user_id = ?", m.ID, m.UserID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("模型不存在或无权限")
	}
	return d.db.WithContext(ctx).Save(m).Error
}

func (d *modelDao) Delete(ctx context.Context, userID uint, id string) error {
	result := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Model{})
	if result.RowsAffected == 0 {
		return errors.New("模型不存在或无权限")
	}
	return result.Error
}

func (d *modelDao) GetByID(ctx context.Context, userID uint, id string) (*model.Model, error) {
	var m model.Model
	err := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模型不存在或无权限")
		}
		return nil, err
	}
	return &m, nil
}

func (d *modelDao) Page(ctx context.Context, userID uint, modelType string, page, size int) ([]*model.Model, int64, error) {
	var models []*model.Model
	var count int64

	db := d.db.WithContext(ctx).Model(&model.Model{}).Where("user_id = ?", userID)
	if modelType != "" {
		db = db.Where("type = ?", modelType)
	}

	err := db.Count(&count).Offset((page - 1) * size).Limit(size).Find(&models).Error
	return models, count, err
}

func (d *modelDao) List(ctx context.Context, userID uint, modelType string) ([]*model.Model, error) {
	var models []*model.Model
	db := d.db.WithContext(ctx).Where("user_id = ?", userID)
	if modelType != "" {
		db = db.Where("type = ?", modelType)
	}
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
