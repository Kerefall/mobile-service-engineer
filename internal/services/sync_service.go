package services

import (
    "context"
    "encoding/base64"
    "fmt"
    "os"
    "strings"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
)

type SyncOrderRequest struct {
    OrderID     int64  `json:"order_id"`
    PhotoBefore string `json:"photo_before,omitempty"`
    PhotoAfter  string `json:"photo_after,omitempty"`
    Signature   string `json:"signature,omitempty"`
    Status      string `json:"status"`
    CompletedAt string `json:"completed_at,omitempty"`
}

type SyncPartItem struct {
    PartID   int64 `json:"part_id"`
    Quantity int   `json:"quantity"`
}

type SyncService struct {
    db          *pgxpool.Pool
    storage     *StorageService
    pdfService  *PDFService
}

func NewSyncService(db *pgxpool.Pool, storage *StorageService, pdfService *PDFService) *SyncService {
    return &SyncService{
        db:         db,
        storage:    storage,
        pdfService: pdfService,
    }
}

func (s *SyncService) SyncOrder(ctx context.Context, orderID int64, photoBefore, photoAfter, signature string) error {
    // Проверяем, не закрыт ли уже заказ
    var status string
    var syncedAt *time.Time
    err := s.db.QueryRow(ctx, "SELECT status, synced_at FROM orders WHERE id = $1", orderID).Scan(&status, &syncedAt)
    if err != nil {
        return fmt.Errorf("ошибка при проверке заказа: %w", err)
    }
    
    if syncedAt != nil {
        logrus.Infof("Заказ %d уже был синхронизирован", orderID)
        return nil
    }
    
    if status == "completed" {
        logrus.Infof("Заказ %d уже закрыт", orderID)
        return nil
    }
    
    // Сохраняем фото и подпись
    var photoBeforePath, photoAfterPath, signaturePath string
    
    if photoBefore != "" {
        photoBeforePath, err = s.saveBase64Image(photoBefore, fmt.Sprintf("order_%d_before", orderID))
        if err != nil {
            logrus.Errorf("Ошибка сохранения фото ДО: %v", err)
        }
    }
    
    if photoAfter != "" {
        photoAfterPath, err = s.saveBase64Image(photoAfter, fmt.Sprintf("order_%d_after", orderID))
        if err != nil {
            logrus.Errorf("Ошибка сохранения фото ПОСЛЕ: %v", err)
        }
    }
    
    if signature != "" {
        signaturePath, err = s.saveBase64Image(signature, fmt.Sprintf("order_%d_signature", orderID))
        if err != nil {
            logrus.Errorf("Ошибка сохранения подписи: %v", err)
        }
    }
    
    // Генерируем PDF
    var pdfPath string
    if photoBeforePath != "" || photoAfterPath != "" {
        pdfPath, err = s.pdfService.GenerateOrderPDF(ctx, orderID, photoBeforePath, photoAfterPath, signaturePath)
        if err != nil {
            logrus.Errorf("Ошибка генерации PDF: %v", err)
        }
    }
    
    // Обновляем заказ
    completedAt := time.Now()
    
    _, err = s.db.Exec(ctx, `
        UPDATE orders SET 
            status = 'completed',
            photo_before_path = COALESCE($1, photo_before_path),
            photo_after_path = COALESCE($2, photo_after_path),
            signature_path = COALESCE($3, signature_path),
            pdf_path = COALESCE($4, pdf_path),
            completed_at = $5,
            synced_at = NOW(),
            updated_at = NOW()
        WHERE id = $6 AND synced_at IS NULL
    `, photoBeforePath, photoAfterPath, signaturePath, pdfPath, completedAt, orderID)
    
    if err != nil {
        return fmt.Errorf("ошибка обновления заказа: %w", err)
    }
    
    logrus.Infof("Заказ %d успешно синхронизирован", orderID)
    return nil
}

func (s *SyncService) saveBase64Image(base64Data string, filename string) (string, error) {
    parts := strings.Split(base64Data, ",")
    if len(parts) > 1 {
        base64Data = parts[1]
    }
    
    imageData, err := base64.StdEncoding.DecodeString(base64Data)
    if err != nil {
        return "", fmt.Errorf("ошибка декодирования base64: %w", err)
    }
    
    // Создаём папку если её нет
    os.MkdirAll("uploads/sync", 0755)
    
    path := fmt.Sprintf("uploads/sync/%s_%d.png", filename, time.Now().Unix())
    err = os.WriteFile(path, imageData, 0644)
    if err != nil {
        return "", fmt.Errorf("ошибка сохранения файла: %w", err)
    }
    
    return "/" + path, nil
}