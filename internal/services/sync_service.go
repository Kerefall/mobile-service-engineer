package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/Kerefall/mobile-service-engineer/internal/dto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type SyncService struct {
	db          *pgxpool.Pool
	storage     *StorageService
	pdfService  *PDFService
	partService *PartService
	onec        *OneCClient
}

func NewSyncService(db *pgxpool.Pool, storage *StorageService, pdfService *PDFService, partService *PartService, onec *OneCClient) *SyncService {
	return &SyncService{
		db:          db,
		storage:     storage,
		pdfService:  pdfService,
		partService: partService,
		onec:        onec,
	}
}

func (s *SyncService) SyncOrder(ctx context.Context, orderID int64, photoBefore, photoAfter, signature string, parts []dto.SyncPartItem) error {
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

	var nParts int
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM order_parts WHERE order_id = $1`, orderID).Scan(&nParts)
	if nParts == 0 && len(parts) > 0 {
		pr := make([]struct {
			PartID   int64
			Quantity int
		}, len(parts))
		for i, p := range parts {
			pr[i].PartID = p.PartID
			pr[i].Quantity = p.Quantity
		}
		if err := s.partService.WriteOffParts(ctx, orderID, pr); err != nil {
			return fmt.Errorf("списание запчастей при синхронизации: %w", err)
		}
	}

	var photoBeforePath, photoAfterPath, signaturePath string

	if photoBefore != "" {
		photoBeforePath, err = s.saveBase64Media(photoBefore, "sync")
		if err != nil {
			logrus.Errorf("Ошибка сохранения фото ДО: %v", err)
		}
	}

	if photoAfter != "" {
		photoAfterPath, err = s.saveBase64Media(photoAfter, "sync")
		if err != nil {
			logrus.Errorf("Ошибка сохранения фото ПОСЛЕ: %v", err)
		}
	}

	if signature != "" {
		signaturePath, err = s.saveBase64Media(signature, "sync")
		if err != nil {
			logrus.Errorf("Ошибка сохранения подписи: %v", err)
		}
	}

	var pdfPath string
	if photoBeforePath != "" || photoAfterPath != "" || signaturePath != "" {
		pdfPath, err = s.pdfService.GenerateOrderPDF(ctx, orderID, photoBeforePath, photoAfterPath, signaturePath)
		if err != nil {
			logrus.Errorf("Ошибка генерации PDF: %v", err)
		}
	}

	completedAt := time.Now()

	_, err = s.db.Exec(ctx, `
        UPDATE orders SET 
            status = 'completed',
            photo_before_path = COALESCE(NULLIF($1, ''), photo_before_path),
            photo_after_path = COALESCE(NULLIF($2, ''), photo_after_path),
            signature_path = COALESCE(NULLIF($3, ''), signature_path),
            pdf_path = COALESCE(NULLIF($4, ''), pdf_path),
            photo_before_at = CASE WHEN $1 <> '' THEN $5::timestamptz ELSE photo_before_at END,
            photo_after_at = CASE WHEN $2 <> '' THEN $5::timestamptz ELSE photo_after_at END,
            completed_at = $5,
            synced_at = NOW(),
            updated_at = NOW()
        WHERE id = $6 AND synced_at IS NULL
    `, photoBeforePath, photoAfterPath, signaturePath, pdfPath, completedAt, orderID)

	if err != nil {
		return fmt.Errorf("ошибка обновления заказа: %w", err)
	}

	logrus.Infof("Заказ %d успешно синхронизирован", orderID)

	order, err := s.loadOrderHead(ctx, orderID)
	if err != nil {
		logrus.Warnf("1С: не удалось загрузить заказ после синхронизации: %v", err)
		return nil
	}
	partLines, err := s.partService.GetPartsLinesForOrder(ctx, orderID)
	if err != nil {
		logrus.Warnf("1С: запчасти: %v", err)
		partLines = nil
	}
	payload := OrderClosePayload{
		OrderID:            order.ID,
		OneCGuid:           order.OneCGuid,
		Title:              order.Title,
		Address:            order.Address,
		Equipment:          order.Equipment,
		CompletedAt:        completedAt,
		PDFWebPath:         order.PDFPath,
		PhotoBeforeWebPath: order.PhotoBeforePath,
		PhotoAfterWebPath:  order.PhotoAfterPath,
		SignatureWebPath:   order.SignaturePath,
		Parts:              partLines,
	}
	if pdfPath != "" {
		payload.PDFWebPath = pdfPath
	}
	if err := s.onec.PushOrderClosed(ctx, payload); err != nil {
		logrus.Warnf("Отправка в 1С после синхронизации: %v", err)
	}

	return nil
}

type orderHead struct {
	ID                 int64
	OneCGuid           string
	Title              string
	Address            string
	Equipment          string
	PDFPath            string
	PhotoBeforePath    string
	PhotoAfterPath     string
	SignaturePath      string
}

func (s *SyncService) loadOrderHead(ctx context.Context, orderID int64) (*orderHead, error) {
	var o orderHead
	err := s.db.QueryRow(ctx, `
		SELECT id, COALESCE(onec_guid,''), title, address, COALESCE(equipment,''),
		       COALESCE(pdf_path,''), COALESCE(photo_before_path,''), COALESCE(photo_after_path,''), COALESCE(signature_path,'')
		FROM orders WHERE id = $1
	`, orderID).Scan(&o.ID, &o.OneCGuid, &o.Title, &o.Address, &o.Equipment, &o.PDFPath, &o.PhotoBeforePath, &o.PhotoAfterPath, &o.SignaturePath)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *SyncService) saveBase64Media(base64Data string, subdir string) (string, error) {
	if i := strings.Index(base64Data, ","); i >= 0 {
		base64Data = base64Data[i+1:]
	}
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования base64: %w", err)
	}
	ext := extFromImageMagic(imageData)
	return s.storage.SaveFile(imageData, subdir, ext)
}

func extFromImageMagic(b []byte) string {
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xD8 {
		return ".jpg"
	}
	if len(b) >= 8 && string(b[0:8]) == "\x89PNG\r\n\x1a\n" {
		return ".png"
	}
	return ".jpg"
}
