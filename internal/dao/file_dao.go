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
// 参数:
//
//	file: 指向要创建的文件模型的指针
//
// 返回值:
//
//	error: 如果创建过程中发生错误，则返回错误信息
func (fd *fileDao) CreateFile(file *model.File) error {
	if fd.db == nil {
		return errors.New("数据库未初始化")
	}
	return fd.db.Create(file).Error
}

// GetFilesByParentID 根据父ID获取文件列表
// 参数:
//
//	userID: 用户ID，用于筛选属于该用户的所有文件
//	parentID: 指向父文件夹的ID，如果为nil，则获取所有顶级文件
//
// 返回值:
//
//	[]model.File: 匹配条件的文件列表
//	error: 如果查询过程中发生错误，则返回错误信息
func (fd *fileDao) GetFilesByParentID(userID uint, parentID *string) ([]model.File, error) {
	var files []model.File
	query := fd.db.Where("user_id = ?", userID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	query = query.Order("is_dir desc")
	query = query.Order("name asc")

	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// GetFileMetaByFileID 根据文件ID获取文件元信息
// 参数:
//
//	id: 文件ID
//
// 返回值:
//
//	*model.File: 如果找到匹配的文件，则返回文件的指针，否则返回nil
//	error: 如果查询过程中发生错误，则返回错误信息
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
// 参数:
//
//	id: 文件ID
//
// 返回值:
//
//	interface{}: 如果删除过程中发生错误，则返回错误信息，否则返回nil
func (fd *fileDao) DeleteFile(id string) error {
	if err := fd.db.Where("id = ?", id).Delete(&model.File{}).Error; err != nil {
		return err
	}
	return nil
}

// ListFiles 列出文件列表，根据指定的排序方式和分页参数
// 参数:
//
//	userID: 用户ID，用于筛选属于该用户的所有文件
//	parentID: 指向父文件夹的ID，如果为nil，则获取所有顶级文件
//	page: 当前页码
//	pageSize: 每页的文件数量
//	sort: 排序参数，格式为"field:order"的字符串，多个排序参数之间用逗号分隔
//
// 返回值:
//
//	[]model.File: 匹配条件的文件列表
//	error: 如果查询过程中发生错误，则返回错误信息
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

// CountFilesByParentID 计算指定父ID下的文件数量
// 参数:
//
//	parentID: 指向父文件夹的ID，如果为nil，则计算所有顶级文件的数量
//	userID: 用户ID，用于筛选属于该用户的所有文件
//
// 返回值:
//
//	int64: 文件数量
//	error: 如果查询过程中发生错误，则返回错误信息
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

// UpdateFile 更新文件信息
// 参数:
//
//	file: 指向要更新的文件模型的指针
//
// 返回值:
//
//	error: 如果更新过程中发生错误，则返回错误信息
func (fd *fileDao) UpdateFile(file *model.File) error {
	if fd.db == nil {
		return errors.New("数据库未初始化")
	}
	return fd.db.Save(file).Error
}
