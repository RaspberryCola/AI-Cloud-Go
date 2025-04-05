package dao

import (
	"ai-cloud/internal/model"
	"fmt"
	"gorm.io/gorm"
)

type KnowledgeBaseDao interface {
	GetDB() *gorm.DB
	// 知识库相关
	CreateKB(kb *model.KnowledgeBase) error                                     // 创建知识库
	DeleteKB(id string) error                                                   // 删除知识库
	CountKBs(userID uint) (int64, error)                                        // 统计知识库数量
	ListKBs(userID uint, page int, pageSize int) ([]model.KnowledgeBase, error) // 获取知识库列表
	GetKBByID(kb_id string) (*model.KnowledgeBase, error)                       // 获取知识库

	// 文档相关
	CreateDocument(doc *model.Document) error                         // 创建文档
	UpdateDocument(doc *model.Document) error                         // 更新文档
	CountDocs(id string) (int64, error)                               // 统计文档数量
	ListDocs(id string, page int, size int) ([]model.Document, error) // 获取文档列表
	GetAllDocsByKBID(kbID string) ([]model.Document, error)           // 获取知识库下所有文档
	DeleteDocsByKBID(kbID string) error                               // 删除知识库下所有文档
}

type kbDao struct {
	db *gorm.DB
}

func NewKnowledgeBaseDao(db *gorm.DB) KnowledgeBaseDao { return &kbDao{db: db} }

func (kd *kbDao) GetDB() *gorm.DB {
	return kd.db
}

func (kd *kbDao) CreateKB(kb *model.KnowledgeBase) error {
	result := kd.db.Create(kb)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (kd *kbDao) GetKBByID(kb_id string) (*model.KnowledgeBase, error) {
	kb := &model.KnowledgeBase{}
	if err := kd.db.Where("id = ?", kb_id).First(kb).Error; err != nil {
		return nil, err
	}
	return kb, nil
}

func (kd *kbDao) CountKBs(userID uint) (int64, error) {
	var total int64
	query := kd.db.Model(&model.KnowledgeBase{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
func (kd *kbDao) ListKBs(userID uint, page int, pageSize int) ([]model.KnowledgeBase, error) {
	var kbs []model.KnowledgeBase
	query := kd.db.Where("user_id = ?", userID).Order("created_at desc")

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	if err := query.Find(&kbs).Error; err != nil {
		return nil, err
	}
	return kbs, nil
}
func (kd *kbDao) CountDocs(kbID string) (int64, error) {
	var total int64
	query := kd.db.Model(&model.Document{}).Where("knowledge_base_id = ?", kbID)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (kd *kbDao) ListDocs(kbID string, page int, size int) ([]model.Document, error) {
	var docs []model.Document
	query := kd.db.Where("knowledge_base_id = ?", kbID).Order("created_at asc")

	offset := (page - 1) * size
	query = query.Offset(offset).Limit(size)
	if err := query.Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}
func (kd *kbDao) DeleteKB(id string) error {
	return kd.db.Where("id = ?", id).Delete(&model.KnowledgeBase{}).Error
}

func (kd *kbDao) CreateDocument(doc *model.Document) error {
	return kd.db.Create(doc).Error
}

func (kd *kbDao) UpdateDocument(doc *model.Document) error {
	if err := kd.db.Save(doc).Error; err != nil {
		return fmt.Errorf("更新文档失败: %w", err)
	}
	return nil
}

func (kd *kbDao) GetAllDocsByKBID(kbID string) ([]model.Document, error) {
	var docs []model.Document
	if err := kd.db.Where("knowledge_base_id = ?", kbID).Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("获取文档失败: %w", err)
	}
	return docs, nil
}

func (kd *kbDao) DeleteDocsByKBID(kbID string) error {
	if err := kd.db.Where("knowledge_base_id = ?", kbID).Delete(&model.Document{}).Error; err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}
	return nil
}
