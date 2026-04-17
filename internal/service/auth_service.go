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

func (s *AuthService) Register(ctx context.Context, name, phone, password string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return s.repo.CreateEngineer(ctx, name, phone, string(hash))
}

func (s *AuthService) Login(ctx context.Context, phone, password string) (string, error) {
	eng, err := s.repo.GetEngineerByPhone(ctx, phone)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(eng.PasswordHash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": eng.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString(s.secret)
}
