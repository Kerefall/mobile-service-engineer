package models

import (
    "time"
)

type Part struct {
    ID              int64     `json:"id"`
    Article         string    `json:"article"`
    Name            string    `json:"name"`
    Description     string    `json:"description"`
    Price           float64   `json:"price"`
    QuantityInStock int       `json:"quantity_in_stock"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

type OrderPart struct {
    ID             int64     `json:"id"`
    OrderID        int64     `json:"order_id"`
    PartID         int64     `json:"part_id"`
    Quantity       int       `json:"quantity"`
    PriceAtMoment  float64   `json:"price_at_moment"`
    CreatedAt      time.Time `json:"created_at"`
}

type WriteOffPartsRequest struct {
    Parts []struct {
        PartID   int64 `json:"part_id" binding:"required"`
        Quantity int   `json:"quantity" binding:"required,min=1"`
    } `json:"parts" binding:"required,min=1"`
}