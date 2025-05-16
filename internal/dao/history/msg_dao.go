package history

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MsgDao interface {
	GetDB() *gorm.DB
	Create(ctx context.Context, msg *model.Message) error
	Update(ctx context.Context, msg *model.Message) error
	Delete(ctx context.Context, msgID string) error
	GetByID(ctx context.Context, msgID string) (*model.Message, error)
	ListByConvID(ctx context.Context, convID string) ([]*model.Message, error)
	List(ctx context.Context, convID string, offset, limit int) ([]*model.Message, int64, error)
	UpdateStatus(ctx context.Context, msgID, status string) error
	UpdateTokenCount(ctx context.Context, msgID string, tokenCount int)
	SetContextEdge(ctx context.Context, msgID string, isContextEdge bool) error
	SetVariant(ctx context.Context, msgID string, isVariant bool) error
}

type msgDao struct {
	db *gorm.DB
}

func NewMsgDao(db *gorm.DB) MsgDao {
	return &msgDao{db: db}
}
func (d *msgDao) GetDB() *gorm.DB {
	return d.db
}

func (d *msgDao) Create(ctx context.Context, msg *model.Message) error {
	if len(msg.MsgID) == 0 {
		msg.MsgID = uuid.NewString()
	}
	err := d.db.WithContext(ctx).Create(msg).Error
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (d *msgDao) Update(ctx context.Context, msg *model.Message) error {
	if err := d.db.WithContext(ctx).Save(msg).Error; err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}
func (d *msgDao) Delete(ctx context.Context, msgID string) error {
	if err := d.db.WithContext(ctx).Delete(&model.Message{}, "msg_id = ?", msgID).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// GetByID 据ID获取消息
func (d *msgDao) GetByID(ctx context.Context, msgID string) (*model.Message, error) {
	var msg model.Message
	if err := d.db.WithContext(ctx).Where("msg_id = ?", msgID).First(&msg).Error; err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return &msg, nil
}

// ListByConvID 根据会话ID获取消息
func (d *msgDao) ListByConvID(ctx context.Context, convID string) ([]*model.Message, error) {
	var msgs []*model.Message
	if err := d.db.WithContext(ctx).Where("conv_id = ?", convID).
		Order("order_seq ASC").
		Find(&msgs).Error; err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	return msgs, nil
}

// Page 获取历史消息
func (d *msgDao) List(ctx context.Context, convID string, offset, limit int) ([]*model.Message, int64, error) {
	var msgs []*model.Message
	var total int64
	db := d.db.WithContext(ctx).Model(&model.Message{}).Where("conv_id = ?", convID).
		Order("order_seq ASC")
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}
	err = db.Offset(offset).Limit(limit).Find(&msgs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages: %w", err)
	}
	return msgs, total, err
}

// UpdateStatus 更新消息状态
func (d *msgDao) UpdateStatus(ctx context.Context, msgID, status string) error {
	err := d.db.WithContext(ctx).Model(&model.Message{}).Where("msg_id = ?", msgID).Update("status", status).Error
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	return nil
}

// UpdateTokenCount 更新消息的token数量
func (d *msgDao) UpdateTokenCount(ctx context.Context, msgID string, tokenCount int) {
	err := d.db.WithContext(ctx).Model(&model.Message{}).Where("msg_id = ?", msgID).Update("token_count", tokenCount).Error
	if err != nil {
		fmt.Printf("failed to update message token count: %v", err)
	}
	return
}

// SetContextEdge 设置消息为上下文边界
func (d *msgDao) SetContextEdge(ctx context.Context, msgID string, isContextEdge bool) error {
	err := d.db.WithContext(ctx).Model(&model.Message{}).Where("msg_id = ?", msgID).Update("is_context_edge", isContextEdge).Error
	if err != nil {
		return fmt.Errorf("failed to update message is_context_edge: %w", err)
	}
	return nil
}

// SetVariant 设置消息为变体消息
func (d *msgDao) SetVariant(ctx context.Context, msgID string, isVariant bool) error {
	err := d.db.WithContext(ctx).Model(&model.Message{}).Where("msg_id = ?", msgID).Update("is_variant", isVariant).Error
	if err != nil {
		return fmt.Errorf("failed to update message is_variant: %w", err)
	}
	return nil
}
