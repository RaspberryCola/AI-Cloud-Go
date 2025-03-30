package dao

import (
	"ai-cloud/internal/model"
	"fmt"
	"gorm.io/gorm"
)

type KnowledgeBaseDao interface {
	CreateKB(kb *model.KnowledgeBase) error
	CountKBs(userID uint) (int64, error)
	ListKBs(userID uint, page int, pageSize int) ([]model.KnowledgeBase, error)
	DeleteKB(id string) error
	CreateDocument(doc *model.Document) error
	UpdateDocument(doc *model.Document) error
	GetKBByID(kb_id string) (*model.KnowledgeBase, error)
}

type kbDao struct {
	db *gorm.DB
}

func NewKnowledgeBaseDao(db *gorm.DB) KnowledgeBaseDao { return &kbDao{db: db} }

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
	query := kd.db.Where("user_id = ?", userID).Order("created_at asc")

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	if err := query.Find(&kbs).Error; err != nil {
		return nil, err
	}
	return kbs, nil
}
func (kd *kbDao) DeleteKB(id string) error {
	return kd.db.Where("id = ?", id).Delete(&model.KnowledgeBase{}).Error
}

func (kd *kbDao) CreateDocument(doc *model.Document) error {
	return kd.db.Create(doc).Error
}

func (kbd *kbDao) UpdateDocument(doc *model.Document) error {
	if err := kbd.db.Save(doc).Error; err != nil {
		return fmt.Errorf("更新文档失败: %w", err)
	}
	return nil
}
