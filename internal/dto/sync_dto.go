package dto

type SyncOrderRequest struct {
    OrderID     int64                     `json:"order_id"`
    PhotoBefore string                    `json:"photo_before,omitempty"` // base64
    PhotoAfter  string                    `json:"photo_after,omitempty"`  // base64
    Signature   string                    `json:"signature,omitempty"`    // base64
    Parts       []SyncPartItem            `json:"parts,omitempty"`
    Status      string                    `json:"status"`
    CompletedAt string                    `json:"completed_at,omitempty"`
}

type SyncPartItem struct {
    PartID   int64 `json:"part_id" binding:"required"`
    Quantity int   `json:"quantity" binding:"required,min=1"`
}

type SyncOrderResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    PDFUrl  string `json:"pdf_url,omitempty"`
}