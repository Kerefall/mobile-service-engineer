package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)

func IdempotencyMiddleware(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.GetHeader("Idempotency-Key")
        if key == "" {
            c.Next()
            return
        }
        
        // Проверяем, не обрабатывался ли уже запрос с таким ключом
        var exists bool
        err := db.QueryRow(c.Request.Context(), 
            "SELECT EXISTS(SELECT 1 FROM idempotency_keys WHERE key = $1)", key).Scan(&exists)
        
        if err == nil && exists {
            // Возвращаем сохранённый ответ
            var responseData map[string]interface{}
            db.QueryRow(c.Request.Context(), 
                "SELECT response_data FROM idempotency_keys WHERE key = $1", key).Scan(&responseData)
            c.JSON(http.StatusOK, responseData)
            c.Abort()
            return
        }
        
        c.Next()
        
        // Сохраняем ответ для будущих запросов
        if c.Writer.Status() == http.StatusOK {
            // Сохраняем responseData в БД
        }
    }
}