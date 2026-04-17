package service

import (
	"fmt"
	"mobile-service-engineer/internal/models"
	"os"

	"github.com/jung-kurt/gofpdf"
)

type PDFService struct{}

func (s *PDFService) GenerateTaskAct(task models.Task, photos []string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Заголовок
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Акт выполненных работ по заказу №%d", task.ID))
	pdf.Ln(12)

	// Инфо о заказе
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 5, fmt.Sprintf("Адрес: %s\nОписание: %s\nСтатус: %s", task.Address, task.Description, task.Status), "", "", false)
	pdf.Ln(10)

	// Вставка фото (если есть)
	if len(photos) > 0 {
		pdf.Cell(40, 10, "Фотофиксация:")
		pdf.Ln(10)
		for _, path := range photos {
			// image, x, y, w, h
			pdf.ImageOptions(path, pdf.GetX(), pdf.GetY(), 50, 0, false, gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}, 0, "")
			pdf.SetX(pdf.GetX() + 55)
		}
	}

	path := fmt.Sprintf("storage/acts/act_%d.pdf", task.ID)
	os.MkdirAll("storage/acts", os.ModePerm)
	err := pdf.OutputFileAndClose(path)
	return path, err
}
