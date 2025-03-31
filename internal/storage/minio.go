package storage

import (
	"ai-cloud/config"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

// NewMinioStorage 创建新的 Minio 存储实例
func NewMinioStorage(cfg config.MinioConfig) (Driver, error) {
	// 设置中国时区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %v", err)
	}
	time.Local = loc

	// 初始化 Minio 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	// 检查 bucket 是否存在
	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	// 如果 bucket 不存在，创建它
	if !exists {
		err = client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return &MinioStorage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload 上传文件到 Minio
func (m *MinioStorage) Upload(data []byte, key string, contentType string) error {
	reader := bytes.NewReader(data)
	_, err := m.client.PutObject(
		context.Background(),
		m.bucket,
		key,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: contentType, // 例如 "application/pdf"
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	return nil
}

// Download 从 Minio 下载文件
func (m *MinioStorage) Download(key string) ([]byte, error) {
	obj, err := m.client.GetObject(
		context.Background(),
		m.bucket,
		key,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %v", err)
	}
	return data, nil
}

// Delete 从 Minio 删除文件
func (m *MinioStorage) Delete(key string) error {
	err := m.client.RemoveObject(context.Background(), m.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}
	return nil
}

func (m *MinioStorage) GetURL(key string) (string, error) {
	// 设置响应头，强制浏览器下载文件
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment")

	// 生成预签名URL，有效期1小时
	expiry := time.Hour * 1
	presignedURL, err := m.client.PresignedGetObject(
		context.Background(),
		m.bucket,
		key,
		expiry,
		reqParams, // 关键：传递自定义参数
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedURL.String(), nil
}

// CreateDirectory 创建目录（通过上传空对象实现）
func (m *MinioStorage) CreateDirectory(dirPath string) error {
	// 确保路径以 / 结尾
	if !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}

	// 上传一个空对象来表示目录
	_, err := m.client.PutObject(context.Background(), m.bucket, dirPath, bytes.NewReader([]byte{}), 0, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	return nil
}
