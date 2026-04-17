package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type respCapture struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *respCapture) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware дедуплицирует запросы с заголовком Idempotency-Key.
// Для /sync order_id берётся из JSON-тела; для /orders/:id/parts — из параметра пути.
func IdempotencyMiddleware(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		var cached []byte
		qerr := db.QueryRow(c.Request.Context(),
			`SELECT response_data FROM idempotency_keys WHERE key = $1`, key).Scan(&cached)
		if qerr == nil && len(cached) > 0 {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			c.Abort()
			return
		}

		orderID, err := resolveIdempotencyOrderID(c)
		if err != nil {
			c.Next()
			return
		}

		cp := &respCapture{ResponseWriter: c.Writer}
		c.Writer = cp

		c.Next()

		if c.Writer.Status() != http.StatusOK || c.IsAborted() {
			return
		}

		body := cp.buf.Bytes()
		if len(body) == 0 {
			return
		}

		_, _ = db.Exec(c.Request.Context(), `
            INSERT INTO idempotency_keys (key, order_id, response_data)
            VALUES ($1, $2, $3::jsonb)
            ON CONFLICT (key) DO NOTHING
        `, key, orderID, body)
	}
}

func resolveIdempotencyOrderID(c *gin.Context) (int64, error) {
	if idStr := c.Param("id"); idStr != "" {
		return strconv.ParseInt(idStr, 10, 64)
	}

	if strings.HasSuffix(c.Request.URL.Path, "/sync") {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return 0, err
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		var meta struct {
			OrderID int64 `json:"order_id"`
		}
		if err := json.Unmarshal(body, &meta); err != nil {
			return 0, err
		}
		return meta.OrderID, nil
	}

	return 0, strconv.ErrSyntax
}
