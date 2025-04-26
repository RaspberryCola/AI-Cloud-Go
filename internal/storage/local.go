package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalStorage 本地存储驱动结构体
type LocalStorage struct {
	baseDir string // 本地存储根目录（如 ./storage_data）
}

// NewLocalStorage 初始化本地存储
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	// 确保存储目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage dir: %v", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

// Upload 上传文件到本地
func (s *LocalStorage) Upload(data []byte, key string, contentType string) error {
	fullPath := filepath.Join(s.baseDir, key)
	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent dir: %v", err)
	}
	return os.WriteFile(fullPath, data, 0644)
}

// Download 从本地下载文件
func (s *LocalStorage) Download(key string) ([]byte, error) {
	fullPath := filepath.Join(s.baseDir, key)
	return os.ReadFile(fullPath)
}

// Delete 删除本地文件
func (s *LocalStorage) Delete(key string) error {
	fullPath := filepath.Join(s.baseDir, key)
	return os.Remove(fullPath)
}

// GetURL 获取本地文件路径（仅返回相对路径）
func (s *LocalStorage) GetURL(key string) (string, error) {
	return filepath.Join(s.baseDir, key), nil
}
