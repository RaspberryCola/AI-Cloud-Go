package dao

import (
	"ai-cloud/internal/model"
	"context"
	"errors"
	"gorm.io/gorm"
)

type AgentDao interface {
	Create(ctx context.Context, agent *model.Agent) error
	Update(ctx context.Context, agent *model.Agent) error
	Delete(ctx context.Context, userID uint, agentID string) error
	GetByID(ctx context.Context, userID uint, agentID string) (*model.Agent, error)
	List(ctx context.Context, userID uint) ([]*model.Agent, error)
	Page(ctx context.Context, userID uint, page, size int) ([]*model.Agent, int64, error)
}

type agentDao struct {
	db *gorm.DB
}

func NewAgentDao(db *gorm.DB) AgentDao {
	return &agentDao{db: db}
}

func (d *agentDao) Create(ctx context.Context, agent *model.Agent) error {
	return d.db.WithContext(ctx).Create(agent).Error
}

func (d *agentDao) Update(ctx context.Context, agent *model.Agent) error {
	// Check if the agent belongs to the user
	var count int64
	if err := d.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ? AND user_id = ?", agent.ID, agent.UserID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("agent not found or no permission")
	}

	// Only update specific fields
	return d.db.WithContext(ctx).Model(agent).
		Select("Name", "Description", "AgentSchema", "UpdatedAt").
		Updates(agent).Error
}

func (d *agentDao) Delete(ctx context.Context, userID uint, agentID string) error {
	result := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", agentID, userID).Delete(&model.Agent{})
	if result.RowsAffected == 0 {
		return errors.New("agent not found or no permission")
	}
	return result.Error
}

func (d *agentDao) GetByID(ctx context.Context, userID uint, agentID string) (*model.Agent, error) {
	var agent model.Agent
	err := d.db.WithContext(ctx).Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found or no permission")
		}
		return nil, err
	}
	return &agent, nil
}

func (d *agentDao) List(ctx context.Context, userID uint) ([]*model.Agent, error) {
	var agents []*model.Agent
	if err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (d *agentDao) Page(ctx context.Context, userID uint, page, size int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var count int64

	db := d.db.WithContext(ctx).Model(&model.Agent{}).Where("user_id = ?", userID)

	err := db.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Offset((page - 1) * size).Limit(size).Find(&agents).Error
	return agents, count, err
}
