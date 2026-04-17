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

func (s *PartService) GetAllParts(ctx context.Context) ([]models.Part, error) {
    rows, err := s.db.Query(ctx, `
        SELECT id, article, name, description, price, quantity_in_stock, created_at, updated_at
        FROM parts ORDER BY name
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var parts []models.Part
    for rows.Next() {
        var part models.Part
        err := rows.Scan(
            &part.ID, &part.Article, &part.Name, &part.Description,
            &part.Price, &part.QuantityInStock, &part.CreatedAt, &part.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        parts = append(parts, part)
    }
    
    return parts, nil
}

func (s *PartService) WriteOffParts(ctx context.Context, orderID int64, parts []struct {
    PartID   int64 `json:"part_id"`
    Quantity int   `json:"quantity"`
}) error {
    for _, p := range parts {
        // Проверяем остаток
        var stock int
        err := s.db.QueryRow(ctx, "SELECT quantity_in_stock FROM parts WHERE id = $1", p.PartID).Scan(&stock)
        if err != nil {
            return fmt.Errorf("запчасть с ID %d не найдена", p.PartID)
        }
        
        if stock < p.Quantity {
            return fmt.Errorf("недостаточно запчасти ID %d на складе: нужно %d, есть %d", p.PartID, p.Quantity, stock)
        }
        
        // Уменьшаем остаток
        _, err = s.db.Exec(ctx, "UPDATE parts SET quantity_in_stock = quantity_in_stock - $1 WHERE id = $2", p.Quantity, p.PartID)
        if err != nil {
            return err
        }
        
        // Добавляем запись о списании
        _, err = s.db.Exec(ctx, `
            INSERT INTO order_parts (order_id, part_id, quantity, price_at_moment)
            SELECT $1, $2, $3, price FROM parts WHERE id = $2
            ON CONFLICT (order_id, part_id) DO UPDATE SET quantity = order_parts.quantity + $3
        `, orderID, p.PartID, p.Quantity)
        if err != nil {
            return err
        }
    }
    
    return nil
}