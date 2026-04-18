package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *pgxpool.Pool
	jwtSecret string
}

func NewAuthService(db *pgxpool.Pool, jwtSecret string) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, login, password, fullName, phone string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля")
	}
	_, err = s.db.Exec(ctx, `
        INSERT INTO engineers (login, password_hash, full_name, phone, is_active)
        VALUES ($1, $2, $3, $4, true)
    `, login, string(hash), fullName, phone)
	if err != nil {
		return fmt.Errorf("не удалось зарегистрировать пользователя")
	}
	return nil
}

type Engineer struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	IsActive bool   `json:"is_active"`
}

// Login возвращает только JWT и ошибку.
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	var id int64
	var dbLogin, passwordHash, fullName, phone string
	var isActive bool

	err := s.db.QueryRow(ctx, `
        SELECT id, login, password_hash, full_name, phone, is_active 
        FROM engineers WHERE login = $1
    `, login).Scan(&id, &dbLogin, &passwordHash, &fullName, &phone, &isActive)
	if err != nil {
		return "", fmt.Errorf("неверный логин или пароль")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("неверный логин или пароль")
	}

	if !isActive {
		return "", fmt.Errorf("учётная запись заблокирована")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    id,
		"login":      login,
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена")
	}

	return tokenString, nil
}

func (s *AuthService) GetEngineerByLogin(ctx context.Context, login string) (*Engineer, error) {
	var id int64
	var dbLogin, fullName, phone string
	var isActive bool

	err := s.db.QueryRow(ctx, `
        SELECT id, login, full_name, phone, is_active 
        FROM engineers WHERE login = $1
    `, login).Scan(&id, &dbLogin, &fullName, &phone, &isActive)
	if err != nil {
		return nil, err
	}

	return &Engineer{
		ID:       id,
		Login:    dbLogin,
		FullName: fullName,
		Phone:    phone,
		IsActive: isActive,
	}, nil
}

func (s *AuthService) GetEngineerByID(ctx context.Context, id int64) (*Engineer, error) {
	var login, fullName, phone string
	var isActive bool

	err := s.db.QueryRow(ctx, `
        SELECT login, full_name, phone, is_active 
        FROM engineers WHERE id = $1
    `, id).Scan(&login, &fullName, &phone, &isActive)
	if err != nil {
		return nil, err
	}

	return &Engineer{
		ID:       id,
		Login:    login,
		FullName: fullName,
		Phone:    phone,
		IsActive: isActive,
	}, nil
}

func (s *AuthService) UpdateFCMToken(ctx context.Context, engineerID int64, fcmToken string) error {
	_, err := s.db.Exec(ctx, `
        UPDATE engineers SET fcm_token = $1, updated_at = NOW() WHERE id = $2
    `, fcmToken, engineerID)
	return err
}
