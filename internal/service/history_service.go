package service

import (
	hisdao "ai-cloud/internal/dao/history"
	"ai-cloud/internal/model"
	"ai-cloud/internal/utils"
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

type HistoryService interface {
	SaveMessage(ctx context.Context, mess *schema.Message, convID string) error
	GetHistory(ctx context.Context, convID string, limit int) ([]*schema.Message, error)
	CreateConversation(ctx context.Context, conv *model.Conversation) error
	UpdateConversation(ctx context.Context, conv *model.Conversation) error
	ArchiveConversation(ctx context.Context, convID string) error
	UnArchiveConversation(ctx context.Context, convID string) error
	PinConversation(ctx context.Context, convID string) error
	UnPinConversation(ctx context.Context, convID string) error
	ListConversations(ctx context.Context, userID uint, page, size int) ([]*model.Conversation, int64, error)
	ListConversationsByAgent(ctx context.Context, userID uint, agentID string, page, size int) ([]*model.Conversation, int64, error)
}
type history struct {
	convDao hisdao.ConvDao
	msgDao  hisdao.MsgDao
}

// NewHistoryService 创建历史记录服务
func NewHistoryService(convDao hisdao.ConvDao, msgDao hisdao.MsgDao) HistoryService {
	return &history{
		convDao: convDao,
		msgDao:  msgDao,
	}
}

// SaveMessage 保存消息
func (s *history) SaveMessage(ctx context.Context, mess *schema.Message, convID string) error {
	err := s.msgDao.Create(ctx, &model.Message{
		Role:    string(mess.Role),
		Content: mess.Content,
		ConvID:  convID,
	})
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

// GetHistory 获取对话的历史消息
func (s *history) GetHistory(ctx context.Context, convID string, limit int) ([]*schema.Message, error) {
	if limit == 0 {
		limit = 50
	}
	_, err := s.convDao.FirstOrCreate(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	msgs, _, err := s.msgDao.List(ctx, convID, 0, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return utils.MessageList2ChatHistory(msgs), nil
}

// CreateConversation 创建会话
func (s *history) CreateConversation(ctx context.Context, conv *model.Conversation) error {
	if err := s.convDao.Create(ctx, conv); err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

// UpdateConversation 更新会话
func (s *history) UpdateConversation(ctx context.Context, conv *model.Conversation) error {
	if err := s.convDao.Update(ctx, conv); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

// ArchiveConversation 归档会话
func (s *history) ArchiveConversation(ctx context.Context, convID string) error {
	if err := s.convDao.Archive(ctx, convID); err != nil {
		return fmt.Errorf("failed to archive conversation: %w", err)
	}
	return nil
}

// UnArchiveConversation 取消归档会话
func (s *history) UnArchiveConversation(ctx context.Context, convID string) error {
	if err := s.convDao.UnArchive(ctx, convID); err != nil {
		return fmt.Errorf("failed to unarchive conversation: %w", err)
	}
	return nil
}

// PinConversation 置顶会话
func (s *history) PinConversation(ctx context.Context, convID string) error {
	if err := s.convDao.Pin(ctx, convID); err != nil {
		return fmt.Errorf("failed to pin conversation: %w", err)
	}
	return nil
}

// UnPinConversation 取消置顶会话
func (s *history) UnPinConversation(ctx context.Context, convID string) error {
	if err := s.convDao.UnPin(ctx, convID); err != nil {
		return fmt.Errorf("failed to unpin conversation: %w", err)
	}
	return nil
}

// ListConversations 获取对话列表
func (s *history) ListConversations(ctx context.Context, userID uint, page, size int) ([]*model.Conversation, int64, error) {
	convs, count, err := s.convDao.Page(ctx, userID, page, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list conversations: %w", err)
	}
	return convs, count, nil
}

// ListConversationsByAgent 按Agent获取对话列表
func (s *history) ListConversationsByAgent(ctx context.Context, userID uint, agentID string, page, size int) ([]*model.Conversation, int64, error) {
	convs, count, err := s.convDao.PageByAgent(ctx, userID, agentID, page, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list conversations by agent: %w", err)
	}
	return convs, count, nil
}
