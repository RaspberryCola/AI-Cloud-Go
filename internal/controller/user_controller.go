package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) Register(ctx *gin.Context) {
	var req model.User

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "用户注册参数错误: "+err.Error())
		return
	}

	if err := c.userService.Register(&req); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "注册失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, req.Username+"注册成功", nil)
}

func (c *UserController) Login(ctx *gin.Context) {
	var req model.UserNameLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "用户名或密码错误")
		return
	}

	loginResponse, err := c.userService.Login(&req)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "登录失败")
		return
	}

	response.SuccessWithMessage(ctx, "登录成功", gin.H{
		"access_token": loginResponse.AccessToken,
		"expires_in":   loginResponse.ExpiresIn,
		"token_type":   loginResponse.TokenType,
	})
}
