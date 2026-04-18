package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kerefall/mobile-service-engineer/internal/config"
)

type StorageService struct {
	uploadPath string
}

func NewStorageService(cfg *config.Config) *StorageService {
	uploadPath := "./uploads"
	os.MkdirAll(uploadPath, 0755)
	os.MkdirAll(filepath.Join(uploadPath, "photos"), 0755)
	os.MkdirAll(filepath.Join(uploadPath, "signatures"), 0755)
	os.MkdirAll(filepath.Join(uploadPath, "pdfs"), 0755)
	os.MkdirAll(filepath.Join(uploadPath, "voice"), 0755)
	
	return &StorageService{uploadPath: uploadPath}
}

func (s *StorageService) SaveFile(data []byte, subdir string, ext string) (string, error) {
	dir := filepath.Join(s.uploadPath, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), subdir, ext)
	filePath := filepath.Join(dir, filename)
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}
	
	return "/static/" + subdir + "/" + filename, nil
}