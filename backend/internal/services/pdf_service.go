package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/sirupsen/logrus"
)

type PDFService struct {
	db *pgxpool.Pool
}

func NewPDFService(db *pgxpool.Pool) *PDFService {
	return &PDFService{db: db}
}

// GenerateOrderPDF генерирует PDF акта: текст, фото до/после, подпись, запчасти.
// Непустые photoBeforePath/photoAfterPath/signaturePath переопределяют пути из БД.
func (s *PDFService) GenerateOrderPDF(ctx context.Context, orderID int64, photoBeforePath, photoAfterPath, signaturePath string) (string, error) {
	var title, address, description, equipment string
	var scheduledDate time.Time
	var dbBefore, dbAfter, dbSig string
	err := s.db.QueryRow(ctx, `
        SELECT title, address, description, COALESCE(equipment,''), scheduled_date,
               photo_before_path, photo_after_path, signature_path
        FROM orders WHERE id = $1
    `, orderID).Scan(&title, &address, &description, &equipment, &scheduledDate, &dbBefore, &dbAfter, &dbSig)
	if err != nil {
		return "", fmt.Errorf("ошибка получения заказа: %w", err)
	}

	before := photoBeforePath
	if strings.TrimSpace(before) == "" {
		before = dbBefore
	}
	after := photoAfterPath
	if strings.TrimSpace(after) == "" {
		after = dbAfter
	}
	sig := signaturePath
	if strings.TrimSpace(sig) == "" {
		sig = dbSig
	}

	rows, err := s.db.Query(ctx, `
        SELECT p.name, op.quantity, op.price_at_moment 
        FROM order_parts op
        JOIN parts p ON op.part_id = p.id
        WHERE op.order_id = $1
    `, orderID)
	if err != nil {
		return "", fmt.Errorf("ошибка получения запчастей: %w", err)
	}
	defer rows.Close()

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Акт выполненных работ")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 8, fmt.Sprintf("Заказ: %s", title), "", "L", false)
	pdf.Ln(2)
	pdf.MultiCell(0, 8, fmt.Sprintf("Адрес: %s", address), "", "L", false)
	if strings.TrimSpace(equipment) != "" {
		pdf.Ln(2)
		pdf.MultiCell(0, 8, fmt.Sprintf("Оборудование: %s", equipment), "", "L", false)
	}
	pdf.Ln(2)
	pdf.MultiCell(0, 8, fmt.Sprintf("Описание: %s", description), "", "L", false)
	pdf.Ln(2)
	pdf.Cell(40, 10, fmt.Sprintf("Дата: %s", scheduledDate.Format("02.01.2006")))
	pdf.Ln(12)

	y := pdf.GetY()
	y = pdfEmbedPhoto(pdf, "Фото «До»", before, y)
	y = pdfEmbedPhoto(pdf, "Фото «После»", after, y)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Использованные запчасти:")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	for rows.Next() {
		var name string
		var quantity int
		var price float64
		if err := rows.Scan(&name, &quantity, &price); err != nil {
			return "", err
		}
		pdf.Cell(40, 8, fmt.Sprintf("- %s x%d = %.2f руб.", name, quantity, price*float64(quantity)))
		pdf.Ln(6)
	}

	vrows, err := s.db.Query(ctx, `SELECT file_path, created_at FROM order_voice_notes WHERE order_id = $1 ORDER BY created_at`, orderID)
	if err != nil {
		logrus.Warnf("PDF: блок голосовых пропущен (%v)", err)
	} else {
		defer vrows.Close()
		headerPrinted := false
		for vrows.Next() {
			if !headerPrinted {
				pdf.Ln(4)
				pdf.SetFont("Arial", "B", 11)
				pdf.Cell(40, 8, "Голосовые сообщения:")
				pdf.Ln(6)
				pdf.SetFont("Arial", "", 9)
				headerPrinted = true
			}
			var vp string
			var vt time.Time
			if err := vrows.Scan(&vp, &vt); err != nil {
				return "", err
			}
			pdf.Cell(0, 6, fmt.Sprintf("- %s (%s)", filepath.Base(vp), vt.Format("02.01.2006 15:04")))
			pdf.Ln(5)
		}
		if err := vrows.Err(); err != nil {
			return "", err
		}
	}

	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 8, "Подпись клиента:")
	pdf.Ln(6)
	sigY := pdf.GetY()
	if fs := localFileFromWebPath(sig); fs != "" {
		opts := gofpdf.ImageOptions{ReadDpi: true}
		pdf.ImageOptions(fs, 12, sigY, 70, 0, false, opts, 0, "")
		pdf.Ln(32)
	} else {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(40, 8, "_________________")
		pdf.Ln(10)
	}

	filename := fmt.Sprintf("order_%d_report_%d.pdf", orderID, time.Now().Unix())
	outputPath := fmt.Sprintf("./uploads/pdfs/%s", filename)
	if err := os.MkdirAll("./uploads/pdfs", 0755); err != nil {
		return "", err
	}
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return "", fmt.Errorf("ошибка сохранения PDF: %w", err)
	}

	return fmt.Sprintf("/static/pdfs/%s", filename), nil
}

func pdfEmbedPhoto(pdf *gofpdf.Fpdf, title, webPath string, y float64) float64 {
	pdf.SetY(y)
	if fs := localFileFromWebPath(webPath); fs == "" {
		return pdf.GetY()
	}
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 8, title)
	pdf.Ln(5)
	opts := gofpdf.ImageOptions{ReadDpi: true}
	fs := localFileFromWebPath(webPath)
	cur := pdf.GetY()
	pdf.ImageOptions(fs, 12, cur, 85, 0, false, opts, 0, "")
	pdf.Ln(58)
	return pdf.GetY()
}

func localFileFromWebPath(webPath string) string {
	webPath = strings.TrimSpace(webPath)
	if webPath == "" {
		return ""
	}
	fs := webPathToFS(webPath)
	if st, err := os.Stat(fs); err != nil || st.IsDir() {
		return ""
	}
	return fs
}
