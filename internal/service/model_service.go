package service

import (
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"context"
)

type ModelService interface {
	CreateModel(ctx context.Context, m *model.Model) error
	UpdateModel(ctx context.Context, m *model.Model) error
	DeleteModel(ctx context.Context, id string) error
	GetModel(ctx context.Context, id string) (*model.Model, error)
	PageModels(ctx context.Context, modelType string, page, size int) ([]*model.Model, int64, error)
}

type modelService struct {
	dao dao.ModelDao
}

func NewModelService(dao dao.ModelDao) ModelService {
	return &modelService{dao: dao}
}

func (s *modelService) CreateModel(ctx context.Context, m *model.Model) error {
	return s.dao.Create(ctx, m)
}

func (s *modelService) UpdateModel(ctx context.Context, m *model.Model) error {
	return s.dao.Update(ctx, m)
}

func (s *modelService) DeleteModel(ctx context.Context, id string) error {
	return s.dao.Delete(ctx, id)
}

func (s *modelService) GetModel(ctx context.Context, id string) (*model.Model, error) {
	return s.dao.GetByID(ctx, id)
}

func (s *modelService) PageModels(ctx context.Context, modelType string, page, size int) ([]*model.Model, int64, error) {
	return s.dao.Page(ctx, modelType, page, size)
}
