package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kerefall/mobile-service-engineer/internal/config"
	"github.com/sirupsen/logrus"
)

// OneCClient отправляет данные закрытого заказа в HTTP-сервис 1С.
type OneCClient struct {
	cfg    *config.Config
	client *http.Client
}

func NewOneCClient(cfg *config.Config) *OneCClient {
	return &OneCClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type oneCPushBody struct {
	Event       string              `json:"event"`
	OrderID     int64               `json:"order_id"`
	OneCGuid    string              `json:"onec_guid,omitempty"`
	Title       string              `json:"title"`
	Address     string              `json:"address"`
	Equipment   string              `json:"equipment,omitempty"`
	CompletedAt string              `json:"completed_at"`
	PDFBase64   string              `json:"pdf_base64,omitempty"`
	Parts       []oneCPartPush      `json:"parts"`
	PhotoBefore string              `json:"photo_before_base64,omitempty"`
	PhotoAfter  string              `json:"photo_after_base64,omitempty"`
	Signature           string   `json:"signature_base64,omitempty"`
	VoiceMessagesBase64 []string `json:"voice_messages_base64,omitempty"`
}

type oneCPartPush struct {
	PartID   int64   `json:"part_id"`
	Article  string  `json:"article"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price_at_moment"`
}

// PushOrderClosed читает PDF и вложения с диска и POSTит JSON на ONEC_HTTP_URL.
func (c *OneCClient) PushOrderClosed(ctx context.Context, p OrderClosePayload) error {
	if c.cfg == nil || strings.TrimSpace(c.cfg.OneCHTTPURL) == "" {
		logrus.Debug("ONEC_HTTP_URL не задан — пропуск отправки в 1С")
		return nil
	}

	pdfB64, err := readFileBase64(p.PDFWebPath)
	if err != nil {
		return fmt.Errorf("чтение PDF для 1С: %w", err)
	}

	body := oneCPushBody{
		Event:       "order_closed",
		OrderID:     p.OrderID,
		OneCGuid:    p.OneCGuid,
		Title:       p.Title,
		Address:     p.Address,
		Equipment:   p.Equipment,
		CompletedAt: p.CompletedAt.UTC().Format(time.RFC3339),
		PDFBase64:   pdfB64,
		Parts:       make([]oneCPartPush, 0, len(p.Parts)),
	}
	for _, x := range p.Parts {
		body.Parts = append(body.Parts, oneCPartPush{
			PartID:   x.PartID,
			Article:  x.Article,
			Name:     x.Name,
			Quantity: x.Quantity,
			Price:    x.Price,
		})
	}
	if b, e := readFileBase64Optional(p.PhotoBeforeWebPath); e == nil {
		body.PhotoBefore = b
	} else if p.PhotoBeforeWebPath != "" {
		logrus.Warnf("1С: фото «до»: %v", e)
	}
	if b, e := readFileBase64Optional(p.PhotoAfterWebPath); e == nil {
		body.PhotoAfter = b
	} else if p.PhotoAfterWebPath != "" {
		logrus.Warnf("1С: фото «после»: %v", e)
	}
	if b, e := readFileBase64Optional(p.SignatureWebPath); e == nil {
		body.Signature = b
	} else if p.SignatureWebPath != "" {
		logrus.Warnf("1С: подпись: %v", e)
	}
	for _, vp := range p.VoiceWebPaths {
		if b, e := readFileBase64Optional(vp); e == nil {
			body.VoiceMessagesBase64 = append(body.VoiceMessagesBase64, b)
		} else if vp != "" {
			logrus.Warnf("1С: голос: %v", e)
		}
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.OneCHTTPURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if t := strings.TrimSpace(c.cfg.OneCHTTPToken); t != "" {
		req.Header.Set("Authorization", "Bearer "+t)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("запрос к 1С: %w", err)
	}
	defer resp.Body.Close()
	slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("1С HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(slurp)))
	}
	logrus.Infof("Заказ #%d отправлен в 1С (HTTP %d)", p.OrderID, resp.StatusCode)
	return nil
}

// OrderClosePayload данные для отправки в 1С при закрытии.
type OrderClosePayload struct {
	OrderID            int64
	OneCGuid           string
	Title              string
	Address            string
	Equipment          string
	CompletedAt        time.Time
	PDFWebPath         string
	PhotoBeforeWebPath string
	PhotoAfterWebPath  string
	SignatureWebPath   string
	VoiceWebPaths      []string
	Parts              []OrderClosePart
}

type OrderClosePart struct {
	PartID   int64
	Article  string
	Name     string
	Quantity int
	Price    float64
}

func readFileBase64Optional(webPath string) (string, error) {
	if strings.TrimSpace(webPath) == "" {
		return "", fmt.Errorf("пустой путь")
	}
	return readFileBase64(webPath)
}

func readFileBase64(webPath string) (string, error) {
	fs := webPathToFS(webPath)
	b, err := os.ReadFile(fs)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// webPathToFS: "/static/photos/x.jpg" -> "./uploads/photos/x.jpg"
func webPathToFS(webPath string) string {
	webPath = strings.TrimPrefix(strings.TrimSpace(webPath), "/")
	if strings.HasPrefix(webPath, "static/") {
		return filepath.Join(".", "uploads", strings.TrimPrefix(webPath, "static/"))
	}
	if strings.HasPrefix(webPath, "uploads/") {
		return filepath.Join(".", webPath)
	}
	return filepath.Join(".", webPath)
}
