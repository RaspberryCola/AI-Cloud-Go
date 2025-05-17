package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"encoding/json"
	"errors"
	"github.com/gin-contrib/sse"
	"io"
	"log"

	"github.com/gin-gonic/gin"
)

type AgentController struct {
	svc service.AgentService
}

func NewAgentController(svc service.AgentService) *AgentController {
	return &AgentController{svc: svc}
}

// CreateAgent 创建agent
func (c *AgentController) CreateAgent(ctx *gin.Context) {
	// Get user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.CreateAgentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	// Create empty agent schema (will be configured during update)
	emptySchema := model.AgentSchema{
		LLMConfig: model.LLMConfig{},
		MCP:       model.MCPConfig{Servers: []string{}},
		Tools:     model.ToolsConfig{ToolIDs: []string{}},
		Prompt:    "",
		Knowledge: model.KnowledgeConfig{KnowledgeIDs: []string{}, TopK: 3},
	}

	// Convert to JSON string
	schemaBytes, err := json.Marshal(emptySchema)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to serialize agent schema")
		return
	}

	// Create new agent with just name and description
	agent := &model.Agent{
		ID:          utils.GenerateUUID(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		AgentSchema: string(schemaBytes),
	}

	if err := c.svc.CreateAgent(ctx.Request.Context(), agent); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to create agent: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "Agent created successfully", gin.H{"id": agent.ID})
}

// UpdateAgent 更新agent
func (c *AgentController) UpdateAgent(ctx *gin.Context) {
	// Get user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.UpdateAgentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	// Get the existing agent
	existingAgent, err := c.svc.GetAgent(ctx.Request.Context(), userID, req.ID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to get agent: "+err.Error())
		return
	}

	// Parse existing schema
	var agentSchema model.AgentSchema
	if err := json.Unmarshal([]byte(existingAgent.AgentSchema), &agentSchema); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to parse existing agent schema")
		return
	}

	// Update name if provided
	if req.Name != "" {
		existingAgent.Name = req.Name
	}

	// Update description if provided
	if req.Description != "" {
		existingAgent.Description = req.Description
	}

	// Update LLMConfig if provided
	if req.LLMConfig.ModelID != "" {
		agentSchema.LLMConfig = req.LLMConfig
	}

	// Update MCP if provided
	if req.MCP.Servers != nil {
		agentSchema.MCP = req.MCP
	}

	// Update Tools if provided
	if req.Tools.ToolIDs != nil {
		agentSchema.Tools = req.Tools
	}

	// Update Prompt if provided
	if req.Prompt != "" {
		agentSchema.Prompt = req.Prompt
	}

	// Update Knowledge if provided
	if req.Knowledge.KnowledgeIDs != nil {
		agentSchema.Knowledge = req.Knowledge
	}

	// Convert updated schema to JSON
	schemaBytes, err := json.Marshal(agentSchema)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to serialize updated agent schema")
		return
	}
	existingAgent.AgentSchema = string(schemaBytes)

	// Update agent
	if err := c.svc.UpdateAgent(ctx.Request.Context(), existingAgent); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to update agent: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "Agent updated successfully", nil)
}

// DeleteAgent 删除agent
func (c *AgentController) DeleteAgent(ctx *gin.Context) {
	// Get user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	agentID := ctx.Query("agent_id")
	if agentID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "Agent ID is required")
		return
	}

	if err := c.svc.DeleteAgent(ctx.Request.Context(), userID, agentID); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to delete agent: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "Agent deleted successfully", nil)
}

// GetAgent handles get agent details requests
func (c *AgentController) GetAgent(ctx *gin.Context) {
	// Get user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	agentID := ctx.Query("agent_id")
	if agentID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "Agent ID is required")
		return
	}

	agent, err := c.svc.GetAgent(ctx.Request.Context(), userID, agentID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to get agent: "+err.Error())
		return
	}

	// Parse agent schema from JSON
	var agentSchema model.AgentSchema
	if err := json.Unmarshal([]byte(agent.AgentSchema), &agentSchema); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to parse agent schema")
		return
	}

	// Return the agent with parsed schema
	response.SuccessWithMessage(ctx, "Agent retrieved successfully", gin.H{
		"id":          agent.ID,
		"user_id":     agent.UserID,
		"name":        agent.Name,
		"description": agent.Description,
		"schema":      agentSchema,
		"created_at":  agent.CreatedAt,
		"updated_at":  agent.UpdatedAt,
	})
}

// PageAgents 分页查询
func (c *AgentController) PageAgents(ctx *gin.Context) {
	// Get user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.PageAgentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	agents, count, err := c.svc.PageAgents(ctx.Request.Context(), userID, req.Page, req.Size)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Failed to get agents: "+err.Error())
		return
	}

	// Prepare response data
	var agentsResponse []gin.H
	for _, agent := range agents {
		// Parse agent schema for each agent
		var agentSchema model.AgentSchema
		if err := json.Unmarshal([]byte(agent.AgentSchema), &agentSchema); err != nil {
			response.InternalError(ctx, errcode.InternalServerError, "Failed to parse agent schema")
			return
		}

		agentsResponse = append(agentsResponse, gin.H{
			"id":          agent.ID,
			"user_id":     agent.UserID,
			"name":        agent.Name,
			"description": agent.Description,
			"schema":      agentSchema,
			"created_at":  agent.CreatedAt,
			"updated_at":  agent.UpdatedAt,
		})
	}

	response.PageSuccess(ctx, agentsResponse, count)
}

// ExecuteAgent 执行Agent
func (c *AgentController) ExecuteAgent(ctx *gin.Context) {
	// Get user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	agentID := ctx.Param("id")
	if agentID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "Agent ID is required")
		return
	}

	var req model.UserMessage
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	result, err := c.svc.ExecuteAgent(ctx.Request.Context(), userID, agentID, req)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Agent execution failed: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "Agent executed successfully", gin.H{"result": result})
}

// StreamExecuteAgent 流式执行Agent
func (ac *AgentController) StreamExecuteAgent(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		log.Printf("[Stream] Error getting user ID: %v\n", err)
		response.UnauthorizedError(c, errcode.UnauthorizedError, "Failed to get user")
		return
	}

	var req model.ExecuteAgentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Stream] Error binding json: %v\n", err)
		response.ParamError(c, errcode.ParamBindError, "Parameter error: "+err.Error())
		return
	}

	sr, err := ac.svc.StreamExecuteAgent(ctx, userID, req.AgentID, req.Message)
	if err != nil {
		log.Printf("[Stream] Error running agent: %v\n", err)
		response.InternalError(c, errcode.InternalServerError, "Agent execution failed: "+err.Error())
		return
	}

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	done := make(chan struct{})
	defer func() {
		sr.Close()
		close(done)
		log.Printf("[Stream] Finish Stream with ID: %s\n", req.ID)
	}()

	// Stream the response
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			log.Printf("[Stream] Context done for chat ID: %s\n", req.ID)
			return false
		case <-done:
			return false
		default:
			msg, err := sr.Recv()
			if errors.Is(err, io.EOF) {
				log.Printf("[Stream] EOF received for chat ID: %s\n", req.ID)
				return false
			}
			if err != nil {
				log.Printf("[Stream] Error receiving message: %v\n", err)
				return false
			}

			// Send SSE event
			sse.Encode(w, sse.Event{
				Data: []byte(msg.Content),
			})

			// Flush the response immediately
			c.Writer.Flush()
			return true
		}
	})
}
