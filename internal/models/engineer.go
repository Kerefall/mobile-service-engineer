package models

import (
    "time"
)

type Engineer struct {
    ID        int64     `json:"id"`
    Login     string    `json:"login"`
    FullName  string    `json:"full_name"`
    Phone     string    `json:"phone"`
    IsActive  bool      `json:"is_active"`
    FCMToken  string    `json:"fcm_token,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type EngineerLoginRequest struct {
    Login    string `json:"login" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type EngineerLoginResponse struct {
    Token string   `json:"token"`
    User  Engineer `json:"user"`
}