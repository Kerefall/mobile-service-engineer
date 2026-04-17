package service

import (
	"context"
	"errors"
	"mobile-service-engineer/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   *repository.TaskRepository
	secret []byte
}

func NewAuthService(repo *repository.TaskRepository, secret string) *AuthService {
	return &AuthService{repo: repo, secret: []byte(secret)}
}

// Register - Регистрация нового инженера
func (s *AuthService) Register(ctx context.Context, fullName, phone, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.CreateEngineer(ctx, fullName, phone, string(hashedPassword))
}

// Login - Проверка данных и выдача токена
func (s *AuthService) Login(ctx context.Context, phone, password string) (string, error) {
	engineer, err := s.repo.GetEngineerByPhone(ctx, phone)
	if err != nil {
		return "", errors.New("пользователь не найден")
	}

	// Сравниваем хэш из БД с введенным паролем
	err = bcrypt.CompareHashAndPassword([]byte(engineer.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("неверный пароль")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": engineer.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString(s.secret)
}
