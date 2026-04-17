package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type FileService struct {
	uploadDir string
}

func NewFileService(uploadDir string) *FileService {
	// Создаем директорию, если её нет
	os.MkdirAll(uploadDir, os.ModePerm)
	return &FileService{uploadDir: uploadDir}
}

func (s *FileService) SaveUpload(file multipart.File, header *multipart.FileHeader, taskID string) (string, error) {
	// Формируем имя: uploads/task123_162543_image.jpg
	filename := fmt.Sprintf("task%s_%d_%s", taskID, time.Now().Unix(), header.Filename)
	filePath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filePath, nil
}