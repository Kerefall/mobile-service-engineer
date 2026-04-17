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

type Order struct {
    ID              int64       `json:"id"`
    Title           string      `json:"title"`
    Description     string      `json:"description"`
    Address         string      `json:"address"`
    Latitude        float64     `json:"latitude,omitempty"`
    Longitude       float64     `json:"longitude,omitempty"`
    ScheduledDate   time.Time   `json:"scheduled_date"`
    Status          OrderStatus `json:"status"`
    EngineerID      int64       `json:"engineer_id"`
    PhotoBeforePath string      `json:"photo_before_path,omitempty"`
    PhotoAfterPath  string      `json:"photo_after_path,omitempty"`
    SignaturePath   string      `json:"signature_path,omitempty"`
    PDFPath         string      `json:"pdf_path,omitempty"`
    ArrivalTime     *time.Time  `json:"arrival_time,omitempty"`
    CompletedAt     *time.Time  `json:"completed_at,omitempty"`
    CreatedAt       time.Time   `json:"created_at"`
    UpdatedAt       time.Time   `json:"updated_at"`
    SyncedAt        *time.Time  `json:"synced_at,omitempty"`
    SyncAttempts    int         `json:"sync_attempts"`
}

// Для ответа с дополнительной информацией
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
    PhotoBefore   string `json:"photo_before,omitempty"`   // base64 фото "До"
    PhotoAfter    string `json:"photo_after,omitempty"`    // base64 фото "После"
    Signature     string `json:"signature,omitempty"`      // base64 подписи
    Notes         string `json:"notes,omitempty"`
}