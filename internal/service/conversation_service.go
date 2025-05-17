package service

import (
	"ai-cloud/internal/model"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

var defaultConvTitle = "新对话"

type ConversationService interface {
	// Debug模式：临时会话，不保存历史
	DebugStreamAgent(ctx context.Context, userID uint, agentID string, message string) (*schema.StreamReader[*schema.Message], error)

	// 会话模式：创建/获取会话，记录历史
	StreamAgentWithConversation(ctx context.Context, userID uint, agentID string, convID string, message string) (*schema.StreamReader[*schema.Message], error)

	// 创建新会话
	CreateConversation(ctx context.Context, userID uint, agentID string) (string, error)

	// 列出用户所有会话
	ListConversations(ctx context.Context, userID uint, page, size int) ([]*model.Conversation, int64, error)

	// 列出特定Agent的会话
	ListAgentConversations(ctx context.Context, userID uint, agentID string, page, size int) ([]*model.Conversation, int64, error)

	// 获取会话历史消息
	GetConversationHistory(ctx context.Context, convID string, limit int) ([]*schema.Message, error)
}

type conversationService struct {
	agentSvc   AgentService
	historySvc HistoryService
}

func NewConversationService(agentSvc AgentService, historySvc HistoryService) ConversationService {
	return &conversationService{
		agentSvc:   agentSvc,
		historySvc: historySvc,
	}
}

// DebugStreamAgent 调试模式：临时会话，不保存历史
func (s *conversationService) DebugStreamAgent(ctx context.Context, userID uint, agentID string, message string) (*schema.StreamReader[*schema.Message], error) {
	// 创建用户消息，不含历史
	userMsg := model.UserMessage{
		Query:   message,
		History: []*schema.Message{},
	}

	// 调用无状态的StreamExecuteAgent
	return s.agentSvc.StreamExecuteAgent(ctx, userID, agentID, userMsg)
}

// StreamAgentWithConversation 会话模式：记录历史
func (s *conversationService) StreamAgentWithConversation(ctx context.Context, userID uint, agentID string, convID string, message string) (*schema.StreamReader[*schema.Message], error) {
	// 确保会话存在
	conv := &model.Conversation{
		ConvID:    convID,
		UserID:    userID,
		AgentID:   agentID,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	// TODO：解决这个问题
	err := s.historySvc.CreateConversation(ctx, conv)
	if err != nil {
		// 可能是会话已存在，忽略错误
		fmt.Printf("创建会话可能已存在: %v\n", err)
	}

	// 先获取历史消息
	historyMsgs, err := s.historySvc.GetHistory(ctx, convID, 50)
	if err != nil {
		return nil, fmt.Errorf("获取历史消息失败: %w", err)
	}

	// 保存用户消息
	userSchemaMsg := &schema.Message{
		Role:    schema.User,
		Content: message,
	}
	err = s.historySvc.SaveMessage(ctx, userSchemaMsg, convID)
	if err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 创建用户消息，包含历史
	userMsg := model.UserMessage{
		Query:   message,
		History: historyMsgs,
	}

	// 调用Agent处理
	sr, err := s.agentSvc.StreamExecuteAgent(ctx, userID, agentID, userMsg)
	if err != nil {
		return nil, err
	}

	// 复制流
	srs := sr.Copy(2)

	// 创建一个独立的上下文用于保存消息，不依赖于请求上下文
	saveCtx := context.Background()

	// 后台处理：记录完整回复
	go func() {
		fullMsgs := make([]*schema.Message, 0)

		defer func() {
			srs[1].Close()
			// 合并消息
			if len(fullMsgs) > 0 {
				fullMsg, err := schema.ConcatMessages(fullMsgs)
				if err != nil {
					fmt.Println("合并消息失败:", err.Error())
					return
				}

				// 使用独立上下文保存消息
				err = s.historySvc.SaveMessage(saveCtx, fullMsg, convID)
				if err != nil {
					fmt.Println("保存消息失败:", err.Error())
				}

				// 更新会话最后更新时间
				conv.UpdatedAt = time.Now().Unix()
				_ = s.historySvc.UpdateConversation(saveCtx, conv)
			}
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				fmt.Println("上下文已关闭:", ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break outer
					}
					fmt.Println("接收消息块错误:", err.Error())
					return
				}

				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()

	return srs[0], nil
}

// CreateConversation 创建新会话
func (s *conversationService) CreateConversation(ctx context.Context, userID uint, agentID string) (string, error) {
	convID := uuid.NewString()

	conv := &model.Conversation{
		ConvID:    convID,
		UserID:    userID,
		AgentID:   agentID,
		Title:     defaultConvTitle,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := s.historySvc.CreateConversation(ctx, conv)
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}

	return convID, nil
}

// ListConversations 列出用户所有会话
func (s *conversationService) ListConversations(ctx context.Context, userID uint, page, size int) ([]*model.Conversation, int64, error) {
	return s.historySvc.ListConversations(ctx, userID, page, size)
}

// ListAgentConversations 列出特定Agent的会话
func (s *conversationService) ListAgentConversations(ctx context.Context, userID uint, agentID string, page, size int) ([]*model.Conversation, int64, error) {
	return s.historySvc.ListConversationsByAgent(ctx, userID, agentID, page, size)
}

// GetConversationHistory 获取会话历史消息
func (s *conversationService) GetConversationHistory(ctx context.Context, convID string, limit int) ([]*schema.Message, error) {
	return s.historySvc.GetHistory(ctx, convID, limit)
}
