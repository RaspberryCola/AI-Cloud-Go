package model

import "time"

type File struct {
	ID          string    `gorm:"primaryKey;type:char(36)"` // UUID
	UserID      uint      `gorm:"index"`                    // 用户ID
	Name        string    `gorm:"not null"`                 // 文件名
	Size        int64     // 文件大小
	Hash        string    `gorm:"index;size:64"` // 文件哈希（SHA-256）
	MIMEType    string    // MIME类型
	IsDir       bool      `gorm:"default:false"`       // 是否为目录
	ParentID    *string   `gorm:"type:char(36);index"` // 父目录ID
	StorageType string    `gorm:"default:'local'"`     // 存储类型：local/oss
	StorageKey  string    // 存储唯一标识（路径或OSS Key）
	CreatedAt   time.Time `gorm:"autoCreateTime"` // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime"` // 更新时间
}

type CreateFolderReq struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

type BatchMoveRequest struct {
	FileIDs        []string `json:"files_pid" binding:"required"`
	TargetParentID string   `json:"target_pid"`
}

type RenameRequest struct {
	FileID  string `json:"file_id" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}
