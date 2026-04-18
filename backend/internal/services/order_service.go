package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Kerefall/mobile-service-engineer/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService struct {
	db *pgxpool.Pool
}

func NewOrderService(db *pgxpool.Pool) *OrderService {
	return &OrderService{db: db}
}

func fillStatusLabel(o *models.Order) {
	o.StatusLabel = o.Status.LabelRu()
}

const orderSelect = `
	id, title, description, COALESCE(equipment,''), COALESCE(onec_guid,''), address, latitude, longitude,
	scheduled_date, status, engineer_id,
	photo_before_path, photo_after_path, photo_before_at, photo_after_at, signature_path, pdf_path,
	arrival_time, completed_at, created_at, updated_at, synced_at, sync_attempts`

func (s *OrderService) GetOrdersByEngineer(ctx context.Context, engineerID int64, status string) ([]models.Order, error) {
	var rows pgx.Rows
	var err error

	if status == "active" {
		rows, err = s.db.Query(ctx, `
            SELECT `+orderSelect+`
            FROM orders 
            WHERE engineer_id = $1 AND status IN ('new', 'in_progress', 'on_the_way')
            ORDER BY scheduled_date
        `, engineerID)
	} else {
		rows, err = s.db.Query(ctx, `
            SELECT `+orderSelect+`
            FROM orders 
            WHERE engineer_id = $1
            ORDER BY scheduled_date
        `, engineerID)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.Title, &order.Description, &order.Equipment, &order.OneCGuid, &order.Address,
			&order.Latitude, &order.Longitude, &order.ScheduledDate,
			&order.Status, &order.EngineerID,
			&order.PhotoBeforePath, &order.PhotoAfterPath, &order.PhotoBeforeAt, &order.PhotoAfterAt, &order.SignaturePath, &order.PDFPath,
			&order.ArrivalTime, &order.CompletedAt, &order.CreatedAt, &order.UpdatedAt, &order.SyncedAt, &order.SyncAttempts,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %v", err)
		}
		fillStatusLabel(&order)
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID int64) (*models.Order, error) {
	var order models.Order
	err := s.db.QueryRow(ctx, `
        SELECT `+orderSelect+`
        FROM orders WHERE id = $1
    `, orderID).Scan(
		&order.ID, &order.Title, &order.Description, &order.Equipment, &order.OneCGuid, &order.Address,
		&order.Latitude, &order.Longitude, &order.ScheduledDate,
		&order.Status, &order.EngineerID,
		&order.PhotoBeforePath, &order.PhotoAfterPath, &order.PhotoBeforeAt, &order.PhotoAfterAt, &order.SignaturePath, &order.PDFPath,
		&order.ArrivalTime, &order.CompletedAt, &order.CreatedAt, &order.UpdatedAt, &order.SyncedAt, &order.SyncAttempts,
	)
	if err != nil {
		return nil, err
	}
	fillStatusLabel(&order)
	return &order, nil
}

func (s *OrderService) GetEngineerIDByLogin(ctx context.Context, login string) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `SELECT id FROM engineers WHERE login = $1 AND is_active = TRUE`, login).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertOrderFromOneC создаёт или обновляет заказ по guid из 1С. isNew=true только при первой вставке (для push).
func (s *OrderService) UpsertOrderFromOneC(ctx context.Context, onecGUID, title, description, equipment, address string, lat, lng float64, scheduledDate time.Time, engineerID int64) (id int64, isNew bool, err error) {
	var existing int64
	err = s.db.QueryRow(ctx, `SELECT id FROM orders WHERE onec_guid = $1`, onecGUID).Scan(&existing)
	if err == nil {
		_, err = s.db.Exec(ctx, `
			UPDATE orders SET
				title = $2, description = $3, equipment = $4, address = $5,
				latitude = $6, longitude = $7, scheduled_date = $8, engineer_id = $9,
				updated_at = NOW()
			WHERE id = $1
		`, existing, title, description, equipment, address, lat, lng, scheduledDate, engineerID)
		if err != nil {
			return 0, false, fmt.Errorf("обновление заказа из 1С: %w", err)
		}
		return existing, false, nil
	}
	if err != pgx.ErrNoRows {
		return 0, false, err
	}

	err = s.db.QueryRow(ctx, `
		INSERT INTO orders (title, description, equipment, address, latitude, longitude, scheduled_date, status, engineer_id, onec_guid, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'new', $8, $9, NOW(), NOW())
		RETURNING id
	`, title, description, equipment, address, lat, lng, scheduledDate, engineerID, onecGUID).Scan(&id)
	if err != nil {
		return 0, false, fmt.Errorf("вставка заказа из 1С: %w", err)
	}
	return id, true, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status models.OrderStatus, lat, lng float64) error {
	var query string
	var err error

	if status == models.StatusOnTheWay {
		query = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
		_, err = s.db.Exec(ctx, query, status, orderID)
	} else if status == models.StatusInProgress {
		query = `UPDATE orders SET status = $1, arrival_time = NOW(), latitude = $2, longitude = $3, updated_at = NOW() WHERE id = $4`
		_, err = s.db.Exec(ctx, query, status, lat, lng, orderID)
	} else {
		query = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
		_, err = s.db.Exec(ctx, query, status, orderID)
	}

	return err
}

func (s *OrderService) CloseOrder(ctx context.Context, orderID int64, pdfPath string) error {
	_, err := s.db.Exec(ctx, `
        UPDATE orders SET 
            status = 'completed', 
            completed_at = NOW(), 
            pdf_path = COALESCE($1, pdf_path),
            updated_at = NOW()
        WHERE id = $2 AND status != 'completed'
    `, pdfPath, orderID)
	return err
}

func (s *OrderService) UpdateOrderPhotos(ctx context.Context, orderID int64, photoBeforePath, photoAfterPath string) error {
	_, err := s.db.Exec(ctx, `
        UPDATE orders SET 
            photo_before_path = CASE WHEN $1 <> '' THEN $1 ELSE photo_before_path END,
            photo_after_path = CASE WHEN $2 <> '' THEN $2 ELSE photo_after_path END,
            photo_before_at = CASE WHEN $1 <> '' THEN NOW() ELSE photo_before_at END,
            photo_after_at = CASE WHEN $2 <> '' THEN NOW() ELSE photo_after_at END,
            updated_at = NOW()
        WHERE id = $3
    `, photoBeforePath, photoAfterPath, orderID)
	return err
}

func (s *OrderService) UpdateOrderSignature(ctx context.Context, orderID int64, signaturePath string) error {
	_, err := s.db.Exec(ctx, `
        UPDATE orders SET signature_path = $1, updated_at = NOW() WHERE id = $2
    `, signaturePath, orderID)
	return err
}

func (s *OrderService) ValidateOrderReadyToClose(ctx context.Context, orderID int64) error {
	order, err := s.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("заказ не найден")
	}

	if order.Status == "completed" {
		return fmt.Errorf("заказ уже закрыт")
	}

	if order.PhotoBeforePath == "" {
		return fmt.Errorf("не загружено фото 'До'")
	}

	if order.PhotoAfterPath == "" {
		return fmt.Errorf("не загружено фото 'После'")
	}

	if order.SignaturePath == "" {
		return fmt.Errorf("нет подписи клиента")
	}

	return nil
}

func (s *OrderService) UpdateOrderPDFPath(ctx context.Context, orderID int64, pdfPath string) error {
	_, err := s.db.Exec(ctx, `
        UPDATE orders SET pdf_path = $1, updated_at = NOW() WHERE id = $2
    `, pdfPath, orderID)
	if err != nil {
		return fmt.Errorf("ошибка обновления пути PDF: %v", err)
	}
	return nil
}

func (s *OrderService) CreateOrder(ctx context.Context, title, description, address string, scheduledDate time.Time, engineerID int64) (int64, error) {
	var orderID int64
	err := s.db.QueryRow(ctx, `
        INSERT INTO orders (title, description, address, scheduled_date, status, engineer_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, 'new', $5, NOW(), NOW())
        RETURNING id
    `, title, description, address, scheduledDate, engineerID).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания заказа: %v", err)
	}
	return orderID, nil
}

func (s *OrderService) AddVoiceNote(ctx context.Context, orderID int64, filePath string, durationSec *int) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO order_voice_notes (order_id, file_path, duration_sec) VALUES ($1, $2, $3) RETURNING id
	`, orderID, filePath, durationSec).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("сохранение голоса: %w", err)
	}
	return id, nil
}

func (s *OrderService) ListVoiceNotes(ctx context.Context, orderID int64) ([]models.VoiceNote, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, order_id, file_path, duration_sec, created_at FROM order_voice_notes WHERE order_id = $1 ORDER BY created_at
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.VoiceNote
	for rows.Next() {
		var v models.VoiceNote
		var dur sql.NullInt64
		if err := rows.Scan(&v.ID, &v.OrderID, &v.FilePath, &dur, &v.CreatedAt); err != nil {
			return nil, err
		}
		if dur.Valid {
			n := int(dur.Int64)
			v.DurationSec = &n
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *OrderService) VoicePathsForOrder(ctx context.Context, orderID int64) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT file_path FROM order_voice_notes WHERE order_id = $1 ORDER BY created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}
