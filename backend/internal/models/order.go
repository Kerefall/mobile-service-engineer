package models

import (
    "time"
)

type OrderStatus string

const (
    StatusNew        OrderStatus = "new"
    StatusInProgress OrderStatus = "in_progress"
    StatusOnTheWay   OrderStatus = "on_the_way"
    StatusCompleted  OrderStatus = "completed"
    StatusSyncing    OrderStatus = "syncing"
)

func (s OrderStatus) IsValid() bool {
    switch s {
    case StatusNew, StatusInProgress, StatusOnTheWay, StatusCompleted, StatusSyncing:
        return true
    }
    return false
}

// LabelRu человекочитаемый статус для UI (MVP: Новое / В работе / Завершено; дополнительно — В пути и др.).
func (s OrderStatus) LabelRu() string {
	switch s {
	case StatusNew:
		return "Новое"
	case StatusInProgress:
		return "В работе"
	case StatusOnTheWay:
		return "В пути"
	case StatusCompleted:
		return "Завершено"
	case StatusSyncing:
		return "Синхронизация"
	default:
		return string(s)
	}
}

type Order struct {
    ID              int64       `json:"id"`
    Title           string      `json:"title"`
    Description     string      `json:"description"`
    Equipment       string      `json:"equipment,omitempty"`
    OneCGuid        string      `json:"onec_guid,omitempty"`
    Address         string      `json:"address"`
    Latitude        float64     `json:"latitude,omitempty"`
    Longitude       float64     `json:"longitude,omitempty"`
    ScheduledDate   time.Time   `json:"scheduled_date"`
    Status          OrderStatus `json:"status"`
    StatusLabel     string      `json:"status_label"`
    EngineerID      int64       `json:"engineer_id"`
    PhotoBeforePath string      `json:"photo_before_path,omitempty"`
    PhotoAfterPath  string      `json:"photo_after_path,omitempty"`
    PhotoBeforeAt   *time.Time  `json:"photo_before_at,omitempty"`
    PhotoAfterAt    *time.Time  `json:"photo_after_at,omitempty"`
    SignaturePath   string      `json:"signature_path,omitempty"`
    PDFPath         string      `json:"pdf_path,omitempty"`
    ArrivalTime     *time.Time  `json:"arrival_time,omitempty"`
    CompletedAt     *time.Time  `json:"completed_at,omitempty"`
    CreatedAt       time.Time   `json:"created_at"`
    UpdatedAt       time.Time   `json:"updated_at"`
    SyncedAt        *time.Time  `json:"synced_at,omitempty"`
    SyncAttempts    int         `json:"sync_attempts"`
}

type OrderWithParts struct {
    Order
    Parts    []OrderPart `json:"parts,omitempty"`
    Engineer *Engineer   `json:"engineer,omitempty"`
}

type UpdateOrderStatusRequest struct {
    Status OrderStatus `json:"status" binding:"required"`
    Lat    float64     `json:"lat,omitempty"`
    Lng    float64     `json:"lng,omitempty"`
}

type CloseOrderRequest struct {
    PhotoBefore string `json:"photo_before,omitempty"`
    PhotoAfter  string `json:"photo_after,omitempty"`
    Signature   string `json:"signature,omitempty"`
    Notes       string `json:"notes,omitempty"`
}