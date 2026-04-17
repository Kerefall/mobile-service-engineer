package service

import (
	"context"
	"mobile-service-engineer/internal/models"
	"mobile-service-engineer/internal/repository"
)

type TaskService struct {
	repo *repository.TaskRepository
}

func NewTaskService(repo *repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// GetTasksForEngineer — здесь мы можем добавить фильтрацию или доп. логику
func (s *TaskService) GetTasksForEngineer(ctx context.Context, engID int) ([]models.Task, error) {
	return s.repo.GetEngineerTasks(ctx, engID)
}

// CloseTask — сложная логика закрытия
func (s *TaskService) CloseTask(ctx context.Context, taskID int, data models.CloseTaskData) error {
	// 1. Обновляем статус в БД
	// 2. Списываем запчасти
	// 3. (Тут будет вызов генерации PDF в будущем)
	return s.repo.FinalizeTask(ctx, taskID, data)
}
