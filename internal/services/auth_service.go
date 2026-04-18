package services

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/golang-jwt/jwt/v5"
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

func (s *AuthService) Login(ctx context.Context, login, password string) (string, *Engineer, error) {
    var id int64
    var dbLogin, passwordHash, fullName, phone string
    var isActive bool
    
    err := s.db.QueryRow(ctx, `
        SELECT id, login, password_hash, full_name, phone, is_active 
        FROM engineers WHERE login = $1
    `, login).Scan(&id, &dbLogin, &passwordHash, &fullName, &phone, &isActive)
    if err != nil {
        return "", nil, fmt.Errorf("неверный логин или пароль")
    }
    
    err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
    if err != nil {
        return "", nil, fmt.Errorf("неверный логин или пароль")
    }
    
    if !isActive {
        return "", nil, fmt.Errorf("учётная запись заблокирована")
    }
    
    // Генерируем JWT токен
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":    id,
        "login":      login,
        "expires_at": time.Now().Add(24 * time.Hour).Unix(),
    })
    
    tokenString, err := token.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", nil, fmt.Errorf("ошибка генерации токена")
    }
    
    engineer := &Engineer{
        ID:       id,
        Login:    dbLogin,
        FullName: fullName,
        Phone:    phone,
        IsActive: isActive,
    }
    
    return tokenString, engineer, nil
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

// GetFCMTokenByEngineerID возвращает сохранённый FCM-токен устройства (для push).
func (s *AuthService) GetFCMTokenByEngineerID(ctx context.Context, engineerID int64) (string, error) {
	var t sql.NullString
	err := s.db.QueryRow(ctx, `SELECT fcm_token FROM engineers WHERE id = $1`, engineerID).Scan(&t)
	if err != nil {
		return "", err
	}
	if !t.Valid {
		return "", nil
	}
	return strings.TrimSpace(t.String), nil
}

// UpdateFCMToken обновляет FCM токен инженера
func (s *AuthService) UpdateFCMToken(ctx context.Context, engineerID int64, fcmToken string) error {
    _, err := s.db.Exec(ctx, `
        UPDATE engineers SET fcm_token = $1, updated_at = NOW() WHERE id = $2
    `, fcmToken, engineerID)
    if err != nil {
        return fmt.Errorf("ошибка обновления FCM токена: %v", err)
    }
    return nil
}