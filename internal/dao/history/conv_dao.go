/*
 * This file contains modified code from the eino-history project.
 * Original source: https://github.com/HildaM/eino-history
 * Licensed under the Apache License, Version 2.0.
 * Modifications are made by RaspberryCola.
 */

package history

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type ConvDao interface {
	Create(ctx context.Context, conv *model.Conversation) error
	Update(ctx context.Context, conv *model.Conversation) error
	Delete(ctx context.Context, convID string) error
	GetByID(ctx context.Context, convID string) (*model.Conversation, error)
	FirstOrCreate(ctx context.Context, conv *model.Conversation) error
	Page(ctx context.Context, userID uint, page, size int) ([]*model.Conversation, int64, error)
	PageByAgent(ctx context.Context, userID uint, agentID string, page, size int) ([]*model.Conversation, int64, error)
	Archive(ctx context.Context, convID string) error
	UnArchive(ctx context.Context, convID string) error
	Pin(ctx context.Context, convID string) error
	UnPin(ctx context.Context, convID string) error
	GetDB() *gorm.DB
}

type convDao struct {
	db *gorm.DB
}

// NewConvDao 创建一个ConvDao
func NewConvDao(db *gorm.DB) ConvDao {
	return &convDao{db: db}
}

func (d *convDao) GetDB() *gorm.DB {
	return d.db
}

// Create 创建一个会话
func (d *convDao) Create(ctx context.Context, conv *model.Conversation) error {
	err := d.db.WithContext(ctx).Create(conv).Error
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

// Update 更新一个会话
func (d *convDao) Update(ctx context.Context, conv *model.Conversation) error {
	err := d.db.WithContext(ctx).Save(conv).Error
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

// Delete 删除一个会话
func (d *convDao) Delete(ctx context.Context, convID string) error {
	err := d.db.WithContext(ctx).Delete(&model.Conversation{}, "conv_id = ?", convID).Error
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

// GetByID 根据ID获取一个会话
func (d *convDao) GetByID(ctx context.Context, convID string) (*model.Conversation, error) {
	var conv model.Conversation
	err := d.db.WithContext(ctx).Where("conv_id = ?", convID).First(&conv).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	return &conv, nil
}

// FirstOrCreate 根据ID获取一个会话，如果会话不存在则创建一个
func (d *convDao) FirstOrCreate(ctx context.Context, conv *model.Conversation) error {
	err := d.db.WithContext(ctx).Where("conv_id = ?", conv.ConvID).FirstOrCreate(&conv).Error
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}
	return nil
}

// Page 分页获取会话
func (d *convDao) Page(ctx context.Context, userID uint, page, size int) ([]*model.Conversation, int64, error) {
	var convs []*model.Conversation
	var total int64

	db := d.db.WithContext(ctx).Model(&model.Conversation{}).Where("user_id = ?", userID).Order("updated_at DESC") // 按照更新时间降序排序

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&convs).Error
	return convs, total, err
}

// PageByAgent 按Agent分页获取会话
func (d *convDao) PageByAgent(ctx context.Context, userID uint, agentID string, page, size int) ([]*model.Conversation, int64, error) {
	var convs []*model.Conversation
	var total int64

	db := d.db.WithContext(ctx).Model(&model.Conversation{}).Where("user_id = ? AND agent_id = ?", userID, agentID).Order("updated_at DESC") // 按照更新时间降序排序

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&convs).Error
	return convs, total, err
}

// Archive 归档一个会话
func (d *convDao) Archive(ctx context.Context, convID string) error {
	err := d.db.WithContext(ctx).Model(&model.Conversation{}).Where("conv_id = ?", convID).Update("is_archived", true).Error
	if err != nil {
		return fmt.Errorf("failed to archive conversation: %w", err)
	}
	return nil
}

// UnArchive 取消归档一个会话
func (d *convDao) UnArchive(ctx context.Context, convID string) error {
	err := d.db.WithContext(ctx).Model(&model.Conversation{}).Where("conv_id = ?", convID).Update("is_archived", false).Error
	if err != nil {
		return fmt.Errorf("failed to unarchive conversation: %w", err)
	}
	return nil
}

// Pin 置顶一个会话
func (d *convDao) Pin(ctx context.Context, convID string) error {
	err := d.db.WithContext(ctx).Model(&model.Conversation{}).Where("conv_id = ?", convID).Update("is_pinned", true).Error
	if err != nil {
		return fmt.Errorf("failed to pin conversation: %w", err)
	}
	return nil
}

// UnPin 取消置顶一个会话
func (d *convDao) UnPin(ctx context.Context, convID string) error {
	err := d.db.WithContext(ctx).Model(&model.Conversation{}).Where("conv_id = ?", convID).Update("is_pinned", false).Error
	if err != nil {
		return fmt.Errorf("failed to unpin conversation: %w", err)
	}
	return nil
}
