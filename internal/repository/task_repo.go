package repository

import (
	"context"
	"database/sql"
	"mobile-service-engineer/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetEngineerTasks выбирает все активные задачи для конкретного инженера
func (r *TaskRepository) GetEngineerTasks(ctx context.Context, engineerID int) ([]models.Task, error) {
	query := `SELECT id, description, address, status, plan_time FROM tasks WHERE engineer_id = $1`
	rows, err := r.db.QueryContext(ctx, query, engineerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Description, &t.Address, &t.Status, &t.PlanTime); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// UpdateTaskStatus обновляет статус и фиксирует координаты
func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, taskID int, status string, lat, lon float64) error {
	query := `UPDATE tasks SET status = $1, last_lat = $2, last_lon = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, status, lat, lon, taskID)
	return err
}

func (r *TaskRepository) CreateEngineer(ctx context.Context, fullName, phone, passwordHash string) error {
	query := `INSERT INTO engineers (full_name, phone, password_hash) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, fullName, phone, passwordHash)
	return err
}
