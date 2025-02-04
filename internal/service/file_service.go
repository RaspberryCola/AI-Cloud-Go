package service

import (
	"ai-cloud/config"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"ai-cloud/internal/storage"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"path/filepath"
	"time"
)

type FileService interface {
	UploadFile(userID uint, fileHeader *multipart.FileHeader, file multipart.File, parentID string) error
	GetFileURL(key string) (string, error)
	PageList(userID uint, parentID *string, page int, pageSize int, sort string) (int64, []model.File, error)
	DownloadFile(fileID string) (*model.File, []byte, error)
	DeleteFileOrFolder(userID uint, fileID string) error
	CreateFolder(userID uint, name string, parentID *string) error
	BatchMoveFiles(userID uint, fileIDs []string, targetParentID string) error
	SearchList(userID uint, key string, page int, size int, sort string) (int64, []model.File, error)
	Rename(userID uint, fileID string, newName string) error
}

type fileService struct {
	fileDao       dao.FileDao
	storageDriver storage.Driver
}

func NewFileService(fileDao dao.FileDao) FileService {
	cfg := config.AppConfigInstance.Storage
	driver, err := storage.NewDriver(cfg)
	if err != nil {
		return nil
	}
	return &fileService{fileDao: fileDao, storageDriver: driver}
}

func (fs *fileService) UploadFile(userID uint, fileHeader *multipart.FileHeader, file multipart.File, parentID string) error {
	fileID := GenerateUUID()
	newFile := model.File{
		ID:          fileID,
		UserID:      userID,
		Name:        fileHeader.Filename,
		Size:        fileHeader.Size,
		MIMEType:    mime.TypeByExtension(filepath.Ext(fileHeader.Filename)),
		StorageType: "local", // TODO: Determine storage type dynamically
		StorageKey:  GenerateStorageKey(userID, fileID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if parentID != "" {
		newFile.ParentID = &parentID
	}
	// TODO:校验ParentID的合法性

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	//
	// Upload file to storage
	if err := fs.storageDriver.Upload(fileData, newFile.StorageKey); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Save file metadata to database
	if err := fs.fileDao.CreateFile(&newFile); err != nil {
		return fmt.Errorf("failed to create file metadata: %w", err)
	}

	return nil
}

func (fs *fileService) GetFileURL(key string) (string, error) {
	return fs.storageDriver.GetURL(key)
}

//func (fs *fileService) ListFiles(userID uint, parentID *string) ([]model.File, error) {
//	return fs.fileDao.GetFilesByParentID(userID, parentID)
//}

func (fs *fileService) CreateFolder(userID uint, name string, parentID *string) error {
	var parent *model.File

	// 父目录验证
	if parentID != nil {
		var err error
		parent, err = fs.fileDao.GetFileMetaByFileID(*parentID)
		if err != nil || parent == nil {
			return errors.New("父目录不存在")
		}
		if !parent.IsDir {
			return errors.New("父路径不是目录")
		}
		if parent.UserID != userID {
			return errors.New("权限不足")
		}
	}

	// 同名检查
	existing, _ := fs.fileDao.GetFilesByParentID(userID, parentID)
	for _, f := range existing {
		if f.Name == name {
			if f.IsDir {
				return errors.New("文件夹已存在")
			}
		}
	}

	// 创建记录
	newFolder := &model.File{
		ID:          GenerateUUID(),
		UserID:      userID,
		Name:        name,
		IsDir:       true,
		ParentID:    parentID,
		StorageType: "dir", // 特殊标识
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := fs.fileDao.CreateFile(newFolder); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

func (fs *fileService) Rename(userID uint, fileID string, newName string) error {

	// 根据id获取file信息
	file, err := fs.fileDao.GetFileMetaByFileID(fileID)
	if err != nil {
		return errors.New("请检查文件是否存在")
	}
	// 检查同名信息
	existing, _ := fs.fileDao.GetFilesByParentID(userID, file.ParentID)
	for _, f := range existing {
		if f.Name == newName {
			return errors.New("目标名称已占用")
		}
	}
	// 更新信息
	file.Name = newName
	file.UpdatedAt = time.Now()
	if err := fs.fileDao.UpdateFile(file); err != nil {
		return errors.New("重命名失败")
	}
	return nil
}
func (fs *fileService) DownloadFile(fileID string) (*model.File, []byte, error) {
	// 1. 验证文件权限并获取元数据
	fileMeta, err := fs.fileDao.GetFileMetaByFileID(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	log.Printf("storagekey为：%s", fileMeta.StorageKey)

	// 2. 从存储驱动获取文件内容
	fileData, err := fs.storageDriver.Download(fileMeta.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("文件下载失败: %w", err)
	}

	// 3. 验证文件大小一致性
	if int64(len(fileData)) != fileMeta.Size {
		return nil, nil, fmt.Errorf("文件大小不匹配")
	}

	return fileMeta, fileData, nil
}

func (fs *fileService) DeleteFileOrFolder(userID uint, fileID string) error {
	file, err := fs.fileDao.GetFileMetaByFileID(fileID)
	if err != nil {
		return fmt.Errorf("获取文件信息失败：%v", err)
	}

	if file.IsDir {
		fileIDPtr := &fileID
		children, err := fs.fileDao.GetFilesByParentID(userID, fileIDPtr)
		if err != nil {
			return fmt.Errorf("获取子文件失败：%v", err)
		}
		for _, child := range children {
			if err := fs.DeleteFileOrFolder(userID, child.ID); err != nil {
				return err
			}
		}
	}
	//删除数据库
	if !file.IsDir {
		storageKey := file.StorageKey
		if err := fs.storageDriver.Delete(storageKey); err != nil {
			return err
		}
	}

	//删除存储
	if err := fs.fileDao.DeleteFile(fileID); err != nil {
		return fmt.Errorf("删除操作失败:%v", err)
	}

	return nil
}

func (fs *fileService) PageList(userID uint, parentID *string, page int, pageSize int, sort string) (int64, []model.File, error) {
	total, err := fs.fileDao.CountFilesByParentID(parentID, userID)
	if err != nil {
		return 0, nil, err
	}
	files, err := fs.fileDao.ListFiles(userID, parentID, page, pageSize, sort)
	if err != nil {
		return 0, nil, err
	}

	return total, files, nil
}

func (fs *fileService) SearchList(userID uint, key string, page int, pageSize int, sort string) (int64, []model.File, error) {

	total, err := fs.fileDao.CountFilesByKeyword(key, userID)
	if err != nil {
		return 0, nil, err
	}
	files, err := fs.fileDao.GetFilesByKeyword(userID, key, page, pageSize, sort)
	if err != nil {
		return 0, nil, err
	}

	return total, files, nil
}

// 批量移动
func (fs *fileService) BatchMoveFiles(userID uint, fileIDs []string, targetParentID string) error {
	// 验证目标文件夹是否存在且合法
	if targetParentID != "" {
		targetFolder, err := fs.fileDao.GetFileMetaByFileID(targetParentID)
		if err != nil || targetFolder == nil {
			return errors.New("目标文件夹不存在")
		}
		if !targetFolder.IsDir {
			return errors.New("目标路径不是文件夹")
		}
		if targetFolder.UserID != userID {
			return errors.New("没有目标文件夹的访问权限")
		}
	}

	// 获取目标文件夹下的所有文件，用于检查同名文件
	var targetParentIDPtr *string
	if targetParentID != "" {
		targetParentIDPtr = &targetParentID
	}
	existingFiles, err := fs.fileDao.GetFilesByParentID(userID, targetParentIDPtr)
	if err != nil {
		return fmt.Errorf("获取目标文件夹内容失败: %w", err)
	}
	existingNames := make(map[string]bool)
	for _, file := range existingFiles {
		existingNames[file.Name] = true
	}

	// 处理每个要移动的文件
	for _, fileID := range fileIDs {
		file, err := fs.fileDao.GetFileMetaByFileID(fileID)
		if err != nil {
			return fmt.Errorf("获取文件信息失败: %w", err)
		}

		// 权限检查
		if file.UserID != userID {
			return errors.New("没有文件的访问权限")
		}

		// 检查是否将文件夹移动到其子文件夹中
		if file.IsDir && targetParentID != "" {
			if err := fs.checkCircularReference(fileID, targetParentID); err != nil {
				return err
			}
		}

		// 处理同名文件
		originalName := file.Name
		newName := originalName
		counter := 1
		for existingNames[newName] {
			ext := filepath.Ext(originalName)
			nameWithoutExt := originalName[:len(originalName)-len(ext)]
			if ext == "" { // 对于文件夹
				newName = fmt.Sprintf("%s (%d)", nameWithoutExt, counter)
			} else { // 对于文件
				newName = fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
			}
			counter++
		}

		// 更新文件信息
		file.Name = newName
		file.ParentID = targetParentIDPtr
		file.UpdatedAt = time.Now()

		if err := fs.fileDao.UpdateFile(file); err != nil {
			return fmt.Errorf("更新文件信息失败: %w", err)
		}

		existingNames[newName] = true
	}

	return nil
}

func (fs *fileService) checkCircularReference(sourceID, targetParentID string) error {
	current := targetParentID
	visited := make(map[string]bool)

	for current != "" {
		if current == sourceID {
			return errors.New("不能将文件夹移动到其子文件夹中")
		}

		if visited[current] {
			return errors.New("检测到文件夹循环引用")
		}
		visited[current] = true

		folder, err := fs.fileDao.GetFileMetaByFileID(current)
		if err != nil {
			return fmt.Errorf("获取文件夹信息失败: %w", err)
		}

		if folder.ParentID == nil {
			break
		}
		current = *folder.ParentID
	}

	return nil
}

func GenerateUUID() string {
	return uuid.New().String()
}
func GenerateStorageKey(userID uint, fileID string) string {
	return fmt.Sprintf("user%d-%s", userID, GenerateUUID())
}
