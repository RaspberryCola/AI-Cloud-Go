package model

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;size:50;not null"`
	Phone     string    `gorm:"uniqueIndex;size:20;not null"` // 新增手机号字段
	Email     string    `gorm:"uniqueIndex;size:100"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=8,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,e164"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}
type UserNameLoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=8,max=30"`
}

type UserPhoneLogin struct {
	Phone    string `json:"phone" binding:"required,e164"`
	Password string `json:"password" binding:"required,min=8,max=30"`
}
