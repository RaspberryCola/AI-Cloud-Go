package dao

import (
	"ai-cloud/internal/model"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

// FileDao 定义了文件操作的接口
type FileDao interface {
	CreateFile(file *model.File) error
	GetFilesByParentID(userID uint, parentID *string) ([]model.File, error)
	GetFileMetaByFileID(id string) (*model.File, error)
	DeleteFile(id string) error
	ListFiles(userID uint, parentID *string, page int, pageSize int, sort string) ([]model.File, error)
	CountFilesByParentID(parentID *string, userID uint) (int64, error)
	UpdateFile(file *model.File) error
	CountFilesByKeyword(key string, userID uint) (int64, error)
	GetFilesByKeyword(userID uint, key string, page int, pageSize int, sort string) ([]model.File, error)
	GetDocumentDir(userID uint) (*model.File, error)
}

// fileDao 实现了FileDao接口，提供文件相关操作
type fileDao struct {
	db *gorm.DB
}

// NewFileDao 创建并返回一个新的FileDao实例
func NewFileDao(db *gorm.DB) FileDao {
	return &fileDao{db: db}
}

// CreateFile 创建一个新的文件记录
func (fd *fileDao) CreateFile(file *model.File) error {
	if fd.db == nil {
		return errors.New("数据库未初始化")
	}
	return fd.db.Create(file).Error
}

// GetFilesByParentID 根据父ID获取文件列表
func (fd *fileDao) GetFilesByParentID(userID uint, parentID *string) ([]model.File, error) {
	var files []model.File
	query := fd.db.Where("user_id = ?", userID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// GetFileMetaByFileID 根据文件ID获取文件元信息
func (fd *fileDao) GetFileMetaByFileID(id string) (*model.File, error) {
	var file model.File
	result := fd.db.Where("id = ?", id).First(&file)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &file, nil
}

// DeleteFile 根据文件ID删除文件记录
func (fd *fileDao) DeleteFile(id string) error {
	if err := fd.db.Where("id = ?", id).Delete(&model.File{}).Error; err != nil {
		return err
	}
	return nil
}

// ListFiles 列出文件列表，根据指定的排序方式和分页参数
func (fd *fileDao) ListFiles(userID uint, parentID *string, page int, pageSize int, sort string) ([]model.File, error) {
	var files []model.File
	query := fd.db.Model(model.File{}).Where("user_id = ?", userID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	query = query.Order("is_dir desc")

	sortClauses := strings.Split(sort, ",")
	for _, clause := range sortClauses {
		parts := strings.Split(clause, ":")
		filed, order := parts[0], parts[1]
		query = query.Order(fmt.Sprintf("%s %s", filed, order))
	}
	//处理分页
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (fd *fileDao) GetFilesByKeyword(userID uint, key string, page int, pageSize int, sort string) ([]model.File, error) {
	var files []model.File
	query := fd.db.Model(&model.File{}).Where("user_id=?", userID).Where("name LIKE ?", "%"+key+"%")

	query = query.Order("is_dir desc")
	sortClauses := strings.Split(sort, ",")
	for _, clause := range sortClauses {
		parts := strings.Split(clause, ":")
		filed, order := parts[0], parts[1]
		query = query.Order(fmt.Sprintf("%s %s", filed, order))
	}
	//处理分页
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// CountFilesByParentID 计算指定父ID下的文件数量
func (fd *fileDao) CountFilesByParentID(parentID *string, userID uint) (int64, error) {
	var total int64
	query := fd.db.Model(&model.File{}).Where("user_id = ?", userID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", parentID)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (fd *fileDao) CountFilesByKeyword(key string, userID uint) (int64, error) {
	var total int64
	query := fd.db.Model(&model.File{}).
		Where("user_id = ?", userID).
		Where("name like ?", "%"+key+"%")
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// UpdateFile 更新文件信息
func (fd *fileDao) UpdateFile(file *model.File) error {
	if fd.db == nil {
		return errors.New("数据库未初始化")
	}
	return fd.db.Save(file).Error
}

func (fd *fileDao) GetDocumentDir(userID uint) (*model.File, error) {
	// 初始化结构体
	file := &model.File{}
	err := fd.db.Where("user_id = ? AND name = ? AND is_dir = ? AND parent_id IS NULL",
		userID, "知识库文件", true).First(file).Error
	if err != nil {
		return nil, err // 直接返回错误，包括 gorm.ErrRecordNotFound
	}
	return file, nil
}
