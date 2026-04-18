package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kerefall/mobile-service-engineer/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
	cfg          *config.Config
	minioClient  *minio.Client
	minioEnabled bool
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
	s := &StorageService{cfg: cfg}

	if cfg.StorageDriver == "minio" {
		mc, err := minio.New(cfg.MinioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
			Secure: cfg.MinioUseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("minio client: %w", err)
		}
		ctx := context.Background()
		err = mc.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			exists, e2 := mc.BucketExists(ctx, cfg.MinioBucket)
			if e2 != nil || !exists {
				return nil, fmt.Errorf("minio bucket %q: %w", cfg.MinioBucket, err)
			}
		}
		s.minioClient = mc
		s.minioEnabled = true
		return s, nil
	}

	_ = os.MkdirAll("./uploads/photos", 0755)
	_ = os.MkdirAll("./uploads/signatures", 0755)
	_ = os.MkdirAll("./uploads/pdfs", 0755)
	_ = os.MkdirAll("./uploads/uploads", 0755)
	_ = os.MkdirAll("./uploads/sync", 0755)

	return s, nil
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

// SaveFile сохраняет байты на диск (/static/...) или в MinIO (публичный URL).
func (s *StorageService) SaveFile(data []byte, folder string, ext string) (publicURL string, err error) {
	if len(data) == 0 {
		return "", fmt.Errorf("пустые данные файла")
	}
	folder = strings.Trim(strings.ReplaceAll(folder, "..", ""), "/")
	ext = normalizeExt(ext)

	if s.minioEnabled && s.minioClient != nil {
		ctx := context.Background()
		name := uniqueFilename("f", ext)
		objectKey := filepath.ToSlash(filepath.Join(folder, name))
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "application/octet-stream"
		}
		_, err = s.minioClient.PutObject(ctx, s.cfg.MinioBucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: ct})
		if err != nil {
			return "", fmt.Errorf("minio put: %w", err)
		}
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.cfg.MinioPublicURL, "/"), s.cfg.MinioBucket, objectKey), nil
	}

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
