package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kerefall/mobile-service-engineer/internal/config"
)

type StorageService struct {
	cfg *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
	_ = os.MkdirAll("./uploads/photos", 0755)
	_ = os.MkdirAll("./uploads/signatures", 0755)
	_ = os.MkdirAll("./uploads/pdfs", 0755)
	_ = os.MkdirAll("./uploads/uploads", 0755)
	_ = os.MkdirAll("./uploads/sync", 0755)

	return &StorageService{cfg: cfg}
}

func normalizeExt(ext string) string {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}

func uniqueFilename(prefix, ext string) string {
	var rnd [8]byte
	_, _ = rand.Read(rnd[:])
	suffix := hex.EncodeToString(rnd[:])
	if prefix == "" {
		prefix = "file"
	}
	return fmt.Sprintf("%s_%d_%s%s", prefix, time.Now().UnixNano(), suffix, normalizeExt(ext))
}

// SaveFile сохраняет байты в ./uploads/{folder}/ с уникальным именем. ext — например ".jpg" или "jpg".
func (s *StorageService) SaveFile(data []byte, folder string, ext string) (publicURL string, err error) {
	if len(data) == 0 {
		return "", fmt.Errorf("пустые данные файла")
	}
	folder = strings.Trim(strings.ReplaceAll(folder, "..", ""), "/")
	dir := filepath.Join("uploads", folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания папки: %w", err)
	}
	name := uniqueFilename("f", ext)
	fullPath := filepath.Join(dir, name)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("ошибка сохранения файла: %w", err)
	}
	rel := filepath.ToSlash(filepath.Join(folder, name))
	return "/static/" + rel, nil
}

// SaveFileReader читает весь r и сохраняет через SaveFile.
func (s *StorageService) SaveFileReader(r io.Reader, folder string, ext string) (publicURL string, err error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения: %w", err)
	}
	return s.SaveFile(data, folder, ext)
}
