package model

import "time"

// BaseModel contains common fields for all models
type BaseModel struct {
    ID        int64     `json:"id" db:"id"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
    Version   int       `json:"version" db:"version"`
}
