package services

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/Kerefall/mobile-service-engineer/internal/models"
)

type PartService struct {
    db *pgxpool.Pool
}

func NewPartService(db *pgxpool.Pool) *PartService {
    return &PartService{db: db}
}

// GetParts возвращает список всех запчастей
func (s *PartService) GetParts(ctx context.Context) ([]models.Part, error) {
    rows, err := s.db.Query(ctx, `
        SELECT id, article, name, description, price, quantity_in_stock 
        FROM parts WHERE quantity_in_stock > 0
        ORDER BY name
    `)
    if err != nil {
        return nil, fmt.Errorf("ошибка получения запчастей: %v", err)
    }
    defer rows.Close()

    var parts []models.Part
    for rows.Next() {
        var p models.Part
        err := rows.Scan(&p.ID, &p.Article, &p.Name, &p.Description, &p.Price, &p.QuantityInStock)
        if err != nil {
            return nil, fmt.Errorf("ошибка сканирования запчасти: %v", err)
        }
        parts = append(parts, p)
    }
    return parts, nil
}

// WriteOffParts списывает запчасти со склада
func (s *PartService) WriteOffParts(ctx context.Context, orderID int64, parts []struct {
    PartID   int64
    Quantity int
}) error {
    for _, p := range parts {
        // Проверяем остаток на складе
        var stock int
        err := s.db.QueryRow(ctx, "SELECT quantity_in_stock FROM parts WHERE id = $1", p.PartID).Scan(&stock)
        if err != nil {
            return fmt.Errorf("запчасть с ID %d не найдена", p.PartID)
        }

        if stock < p.Quantity {
            return fmt.Errorf("недостаточно запчастей на складе (ID: %d, доступно: %d, нужно: %d)", p.PartID, stock, p.Quantity)
        }

        // Уменьшаем остаток
        _, err = s.db.Exec(ctx, "UPDATE parts SET quantity_in_stock = quantity_in_stock - $1 WHERE id = $2", p.Quantity, p.PartID)
        if err != nil {
            return fmt.Errorf("ошибка обновления остатка: %v", err)
        }

        // Добавляем запись о списании
        _, err = s.db.Exec(ctx, `
            INSERT INTO order_parts (order_id, part_id, quantity, price_at_moment) 
            SELECT $1, $2, $3, price FROM parts WHERE id = $2
        `, orderID, p.PartID, p.Quantity)
        if err != nil {
            return fmt.Errorf("ошибка записи списания: %v", err)
        }
    }
    return nil
}