package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"errors"
	"io"
	"log"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConversationController struct {
	svc service.ConversationService
}

func NewConversationController(svc service.ConversationService) *ConversationController {
	return &ConversationController{svc: svc}
}

// DebugStreamAgent 调试模式，不保存历史
func (c *ConversationController) DebugStreamAgent(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.DebugRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	// 调用debug模式流式处理
	sr, err := c.svc.DebugStreamAgent(ctx.Request.Context(), userID, req.AgentID, req.Message)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Agent execution failed: "+err.Error())
		return
	}

	// 设置SSE响应头
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 传输流
	sessionID := uuid.NewString()
	done := make(chan struct{})
	defer func() {
		sr.Close()
		close(done)
		log.Printf("[Debug Stream] Finish Stream with ID: %s\n", sessionID)
	}()

	// 流式响应
	ctx.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Request.Context().Done():
			log.Printf("[Debug Stream] Context done for session ID: %s\n", sessionID)
			return false
		case <-done:
			return false
		default:
			msg, err := sr.Recv()
			if errors.Is(err, io.EOF) {
				log.Printf("[Debug Stream] EOF received for session ID: %s\n", sessionID)
				return false
			}
			if err != nil {
				log.Printf("[Debug Stream] Error receiving message: %v\n", err)
				return false
			}

			// 发送SSE事件
			sse.Encode(w, sse.Event{
				Data: []byte(msg.Content),
			})

			// 立即刷新响应
			ctx.Writer.Flush()
			return true
		}
	})
}

// CreateConversation 创建新会话
func (c *ConversationController) CreateConversation(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.CreateConvRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	// 创建会话
	convID, err := c.svc.CreateConversation(ctx.Request.Context(), userID, req.AgentID)
	if err != nil {
		log.Printf("[Create] Error creating conversation: %v\n", err)
		response.InternalError(ctx, errcode.InternalServerError, "Failed to create conversation")
		return
	}

	response.SuccessWithMessage(ctx, "Conversation created successfully", gin.H{"conv_id": convID})
}

// StreamConversation 会话模式，保存历史
func (c *ConversationController) StreamConversation(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.ConvRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	// 调用会话模式流式处理
	sr, err := c.svc.StreamAgentWithConversation(ctx.Request.Context(), userID, req.AgentID, req.ConvID, req.Message)
	if err != nil {
		log.Printf("[Conversation Stream] Error running agent: %v\n", err)
		response.InternalError(ctx, errcode.InternalServerError, "Agent execution failed")
		return
	}

	// 设置SSE响应头
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 传输流
	done := make(chan struct{})
	defer func() {
		sr.Close()
		close(done)
		log.Printf("[Conversation Stream] Finish Stream with ConvID: %s\n", req.ConvID)
	}()

	// 流式响应
	ctx.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Request.Context().Done():
			log.Printf("[Conversation Stream] Context done for ConvID: %s\n", req.ConvID)
			return false
		case <-done:
			return false
		default:
			msg, err := sr.Recv()
			if errors.Is(err, io.EOF) {
				log.Printf("[Conversation Stream] EOF received for ConvID: %s\n", req.ConvID)
				return false
			}
			if err != nil {
				log.Printf("[Conversation Stream] Error receiving message: %v\n", err)
				return false
			}

			// 发送SSE事件
			sse.Encode(w, sse.Event{
				Data: []byte(msg.Content),
			})

			// 立即刷新响应
			ctx.Writer.Flush()
			return true
		}
	})
}

// ListConversations 获取用户所有会话
func (c *ConversationController) ListConversations(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	// 分页参数
	page := utils.StringToInt(ctx.DefaultQuery("page", "1"))
	size := utils.StringToInt(ctx.DefaultQuery("size", "10"))

	// 获取会话列表
	convs, count, err := c.svc.ListConversations(ctx.Request.Context(), userID, page, size)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to list conversations: "+err.Error())
		return
	}

	// 返回分页数据
	response.PageSuccess(ctx, convs, count)
}

// ListAgentConversations 获取特定Agent的会话
func (c *ConversationController) ListAgentConversations(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	// 获取AgentID
	agentID := ctx.Query("agent_id")
	if agentID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "Agent ID is required")
		return
	}

	// 分页参数
	page := utils.StringToInt(ctx.DefaultQuery("page", "1"))
	size := utils.StringToInt(ctx.DefaultQuery("size", "10"))

	// 获取会话列表
	convs, count, err := c.svc.ListAgentConversations(ctx.Request.Context(), userID, agentID, page, size)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to list agent conversations: "+err.Error())
		return
	}

	// 返回分页数据
	response.PageSuccess(ctx, convs, count)
}

// GetConversationHistory 获取会话历史消息
func (c *ConversationController) GetConversationHistory(ctx *gin.Context) {
	// 获取会话ID
	convID := ctx.Query("conv_id")
	if convID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "Conversation ID is required")
		return
	}

	// 限制参数
	limit := utils.StringToInt(ctx.DefaultQuery("limit", "50"))

	// 获取历史消息
	msgs, err := c.svc.GetConversationHistory(ctx.Request.Context(), convID, limit)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to get conversation history: "+err.Error())
		return
	}

	// 返回历史消息
	response.SuccessWithMessage(ctx, "Conversation history retrieved successfully", gin.H{"messages": msgs})
}
