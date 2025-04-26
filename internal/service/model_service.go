package service

import (
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"context"
)

type ModelService interface {
	CreateModel(ctx context.Context, m *model.Model) error
	UpdateModel(ctx context.Context, m *model.Model) error
	DeleteModel(ctx context.Context, userID uint, id string) error
	GetModel(ctx context.Context, userID uint, id string) (*model.Model, error)
	ListModels(ctx context.Context, userID uint, modelType string) ([]*model.Model, error)
	PageModels(ctx context.Context, userID uint, modelType string, page, size int) ([]*model.Model, int64, error)
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

func (s *modelService) DeleteModel(ctx context.Context, userID uint, id string) error {
	return s.dao.Delete(ctx, userID, id)
}

func (s *modelService) GetModel(ctx context.Context, userID uint, id string) (*model.Model, error) {
	return s.dao.GetByID(ctx, userID, id)
}

func (s *modelService) ListModels(ctx context.Context, userID uint, modelType string) ([]*model.Model, error) {
	return s.dao.List(ctx, userID, modelType)
}

func (s *modelService) PageModels(ctx context.Context, userID uint, modelType string, page, size int) ([]*model.Model, int64, error) {
	return s.dao.Page(ctx, userID, modelType, page, size)
}
