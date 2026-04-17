package service

import (
	"context"
	"fmt"
	"strings"
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

func (s *OrderService) GetOrdersByEngineer(ctx context.Context, engineerID int64, status string) ([]models.Order, error) {
	var rows pgx.Rows
	var err error

	if status == "active" {
		rows, err = s.db.Query(ctx, `
            SELECT id, title, description, address, latitude, longitude, 
                   scheduled_date, status, engineer_id, created_at, updated_at
            FROM orders 
            WHERE engineer_id = $1 AND status IN ('new', 'in_progress', 'on_the_way')
            ORDER BY scheduled_date
        `, engineerID)
	} else {
		rows, err = s.db.Query(ctx, `
            SELECT id, title, description, address, latitude, longitude, 
                   scheduled_date, status, engineer_id, created_at, updated_at
            FROM orders 
            WHERE engineer_id = $1
            ORDER BY scheduled_date
        `, engineerID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.Title, &order.Description, &order.Address,
			&order.Latitude, &order.Longitude, &order.ScheduledDate,
			&order.Status, &order.EngineerID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID int64) (*models.Order, error) {
	var order models.Order
	err := s.db.QueryRow(ctx, `
        SELECT id, title, description, address, latitude, longitude, 
               scheduled_date, status, engineer_id, 
               photo_before_path, photo_after_path, signature_path, pdf_path,
               arrival_time, completed_at, created_at, updated_at
        FROM orders WHERE id = $1
    `, orderID).Scan(
		&order.ID, &order.Title, &order.Description, &order.Address,
		&order.Latitude, &order.Longitude, &order.ScheduledDate,
		&order.Status, &order.EngineerID,
		&order.PhotoBeforePath, &order.PhotoAfterPath, &order.SignaturePath, &order.PDFPath,
		&order.ArrivalTime, &order.CompletedAt, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ValidateOrderReadyToClose проверяет фото до/после и подпись (запчасти по ТЗ опциональны).
func (s *OrderService) ValidateOrderReadyToClose(ctx context.Context, orderID int64) error {
	o, err := s.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	var missing []string
	if strings.TrimSpace(o.PhotoBeforePath) == "" {
		missing = append(missing, "фото «до»")
	}
	if strings.TrimSpace(o.PhotoAfterPath) == "" {
		missing = append(missing, "фото «после»")
	}
	if strings.TrimSpace(o.SignaturePath) == "" {
		missing = append(missing, "подпись")
	}
	if len(missing) > 0 {
		return fmt.Errorf("нельзя закрыть заказ: не заполнены %s", strings.Join(missing, ", "))
	}
	return nil
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
            photo_before_path = COALESCE($1, photo_before_path),
            photo_after_path = COALESCE($2, photo_after_path),
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

func (s *OrderService) UpdateOrderPDFPath(ctx context.Context, orderID int64, pdfPath string) error {
	_, err := s.db.Exec(ctx, `UPDATE orders SET pdf_path = $1, updated_at = NOW() WHERE id = $2`, pdfPath, orderID)
	return err
}

// CreateOrder создаёт заказ и возвращает id (для админки / демо).
func (s *OrderService) CreateOrder(ctx context.Context, title, description, address string, engineerID int64, lat, lng float64, scheduled time.Time) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
        INSERT INTO orders (title, description, address, latitude, longitude, scheduled_date, status, engineer_id)
        VALUES ($1, $2, $3, $4, $5, $6, 'new', $7)
        RETURNING id
    `, title, description, address, lat, lng, scheduled, engineerID).Scan(&id)
	return id, err
}

func (s *OrderService) OrderBelongsToEngineer(ctx context.Context, orderID, engineerID int64) (bool, error) {
	var eng int64
	err := s.db.QueryRow(ctx, `SELECT engineer_id FROM orders WHERE id = $1`, orderID).Scan(&eng)
	if err != nil {
		return false, err
	}
	return eng == engineerID, nil
}
