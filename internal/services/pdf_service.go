package services

import (
    "context"
    "fmt"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jung-kurt/gofpdf/v2"
)

type PDFService struct {
    db *pgxpool.Pool
}

func NewPDFService(db *pgxpool.Pool) *PDFService {
    return &PDFService{db: db}
}

// GenerateOrderPDF генерирует PDF акт для заказа
func (s *PDFService) GenerateOrderPDF(ctx context.Context, orderID int64, photoBeforePath, photoAfterPath, signaturePath string) (string, error) {
    // Получаем данные заказа
    var title, address, description string
    var scheduledDate time.Time
    err := s.db.QueryRow(ctx, `
        SELECT title, address, description, scheduled_date 
        FROM orders WHERE id = $1
    `, orderID).Scan(&title, &address, &description, &scheduledDate)
    if err != nil {
        return "", fmt.Errorf("ошибка получения заказа: %w", err)
    }
    
    // Получаем запчасти
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
    
    // Создаём PDF
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(40, 10, "Акт выполненных работ")
    pdf.Ln(12)
    
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(40, 10, fmt.Sprintf("Заказ: %s", title))
    pdf.Ln(8)
    pdf.Cell(40, 10, fmt.Sprintf("Адрес: %s", address))
    pdf.Ln(8)
    pdf.Cell(40, 10, fmt.Sprintf("Описание: %s", description))
    pdf.Ln(8)
    pdf.Cell(40, 10, fmt.Sprintf("Дата: %s", scheduledDate.Format("02.01.2006")))
    pdf.Ln(12)
    
    pdf.SetFont("Arial", "B", 12)
    pdf.Cell(40, 10, "Использованные запчасти:")
    pdf.Ln(8)
    
    pdf.SetFont("Arial", "", 10)
    for rows.Next() {
        var name string
        var quantity int
        var price float64
        rows.Scan(&name, &quantity, &price)
        pdf.Cell(40, 8, fmt.Sprintf("- %s x%d = %.2f руб.", name, quantity, price*float64(quantity)))
        pdf.Ln(6)
    }
    
    pdf.Ln(10)
    pdf.Cell(40, 10, "Подпись клиента: _________________")
    
    // Сохраняем PDF
    filename := fmt.Sprintf("order_%d_report_%d.pdf", orderID, time.Now().Unix())
    outputPath := fmt.Sprintf("./uploads/pdfs/%s", filename)
    err = pdf.OutputFileAndClose(outputPath)
    if err != nil {
        return "", fmt.Errorf("ошибка сохранения PDF: %w", err)
    }
    
    return fmt.Sprintf("/static/pdfs/%s", filename), nil
}