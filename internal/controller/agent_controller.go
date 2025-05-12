package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

type AgentController struct {
	svc service.AgentService
}

func NewAgentController(svc service.AgentService) *AgentController {
	return &AgentController{svc: svc}
}

// CreateAgent handles initial agent creation with just name and description
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
		Knowledge: model.KnowledgeConfig{KnowledgeIDs: []string{}},
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

// UpdateAgent handles full agent configuration update
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

// DeleteAgent handles agent deletion requests
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

// PageAgents handles paginated agent list requests
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

// ExecuteAgent handles agent execution requests
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

// StreamExecuteAgent handles streaming agent execution requests
func (c *AgentController) StreamExecuteAgent(ctx *gin.Context) {
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

	msgReader, err := c.svc.StreamExecuteAgent(ctx.Request.Context(), userID, agentID, req)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "Agent execution failed: "+err.Error())
		return
	}

	// Set headers for SSE
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	ctx.Writer.Flush()

	// Stream the response
	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Reached end of stream
				break
			}
			// Handle error
			ctx.Writer.WriteString("event: error\n")
			ctx.Writer.WriteString(fmt.Sprintf("data: %s\n\n", err.Error()))
			ctx.Writer.Flush()
			return
		}

		// Write message to client
		ctx.Writer.WriteString("event: message\n")
		ctx.Writer.WriteString(fmt.Sprintf("data: %s\n\n", msg.Content))
		ctx.Writer.Flush()
	}

	// Signal end of stream
	ctx.Writer.WriteString("event: done\n")
	ctx.Writer.WriteString("data: \n\n")
	ctx.Writer.Flush()
}
