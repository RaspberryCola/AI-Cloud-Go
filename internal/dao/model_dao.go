package dao

import (
	"ai-cloud/internal/model"
	"context"
	"gorm.io/gorm"
)

type ModelDao interface {
	Create(ctx context.Context, m *model.Model) error
	Update(ctx context.Context, m *model.Model) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.Model, error)
	Page(ctx context.Context, modelType string, page, size int) ([]*model.Model, int64, error)
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
	return d.db.WithContext(ctx).Save(m).Error
}

func (d *modelDao) Delete(ctx context.Context, id string) error {
	return d.db.WithContext(ctx).Delete(&model.Model{ID: id}).Error
}

func (d *modelDao) GetByID(ctx context.Context, id string) (*model.Model, error) {
	var m model.Model
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	return &m, err
}

func (d *modelDao) Page(ctx context.Context, modelType string, page, size int) ([]*model.Model, int64, error) {
	var models []*model.Model
	var count int64

	db := d.db.WithContext(ctx).Model(&model.Model{})
	if modelType != "" {
		db = db.Where("type = ?", modelType)
	}

	err := db.Count(&count).Offset((page - 1) * size).Limit(size).Find(&models).Error
	return models, count, err
}

func (d *modelDao) List(ctx context.Context, modelType string) ([]*model.Model, error) {
	var models []*model.Model
	if err := d.db.WithContext(ctx).Where("type = ?", modelType).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
