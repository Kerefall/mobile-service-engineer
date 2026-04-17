package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func AuthMiddleware(jwtSecret string, db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "отсутствует токен авторизации"})
            c.Abort()
            return
        }
        
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
            c.Abort()
            return
        }
        
        tokenString := parts[1]
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен"})
            c.Abort()
            return
        }
        
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен"})
            c.Abort()
            return
        }
        
        userID, ok := claims["user_id"].(float64)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен"})
            c.Abort()
            return
        }
        
        c.Set("user_id", int64(userID))
        c.Set("login", claims["login"])
        c.Next()
    }
}