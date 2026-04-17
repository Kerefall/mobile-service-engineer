package services

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
    "github.com/Kerefall/mobile-service-engineer/internal/config"
)

type StorageService struct {
    cfg *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
    // Создаём папку для загрузок если её нет
    os.MkdirAll("./uploads/photos", 0755)
    os.MkdirAll("./uploads/signatures", 0755)
    os.MkdirAll("./uploads/pdfs", 0755)
    
    return &StorageService{cfg: cfg}
}

// SaveFile сохраняет файл на диск и возвращает путь
func (s *StorageService) SaveFile(data []byte, folder, filename string) (string, error) {
    dir := fmt.Sprintf("./uploads/%s", folder)
    err := os.MkdirAll(dir, 0755)
    if err != nil {
        return "", fmt.Errorf("ошибка создания папки: %w", err)
    }
    
    fullPath := filepath.Join(dir, filename)
    err = os.WriteFile(fullPath, data, 0644)
    if err != nil {
        return "", fmt.Errorf("ошибка сохранения файла: %w", err)
    }
    
    return fmt.Sprintf("/static/%s/%s", folder, filename), nil
}

// GenerateUniqueFilename генерирует уникальное имя файла
func (s *StorageService) GenerateUniqueFilename(prefix, ext string) string {
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("%s_%d%s", prefix, timestamp, ext)
}