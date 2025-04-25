package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/pkgs/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ModelController struct {
	svc service.ModelService
}

func NewModelController(svc service.ModelService) *ModelController {
	return &ModelController{svc: svc}
}

func (c *ModelController) CreateModel(ctx *gin.Context) {
	var req model.CreateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &model.Model{
		Type:    req.Type,
		Name:    req.Name,
		Server:  req.Server,
		BaseURL: req.BaseURL,
		Model:   req.Model,
		APIKey:  req.APIKey,
		// embedding
		Dimension: req.Dimension,
		// llm
		MaxOutputLength: req.MaxOutputLength,
		Function:        req.Function,
		// common
		MaxTokens: req.MaxTokens,
	}

	if err := c.svc.CreateModel(ctx.Request.Context(), m); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, m)
}

func (c *ModelController) UpdateModel(ctx *gin.Context) {
	id := ctx.Param("id")
	var req model.CreateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &model.Model{
		ID:      id,
		Type:    req.Type,
		Name:    req.Name,
		Server:  req.Server,
		BaseURL: req.BaseURL,
		Model:   req.Model,
		APIKey:  req.APIKey,
		// embedding
		Dimension: req.Dimension,
		// llm
		MaxOutputLength: req.MaxOutputLength,
		Function:        req.Function,
		// common
		MaxTokens: req.MaxTokens,
	}

	if err := c.svc.UpdateModel(ctx.Request.Context(), m); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, m)
}

func (c *ModelController) DeleteModel(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.DeleteModel(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (c *ModelController) GetModel(ctx *gin.Context) {
	id := ctx.Param("id")
	m, err := c.svc.GetModel(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, m)
}

func (c *ModelController) PageModels(ctx *gin.Context) {
	var req struct {
		Type string `form:"type"`
		Page int    `form:"page,default=1"`
		Size int    `form:"size,default=10"`
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models, count, err := c.svc.PageModels(ctx.Request.Context(), req.Type, req.Page, req.Size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response.PageSuccess(ctx, models, count)
}
