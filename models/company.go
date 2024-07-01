package models

import (
    "time"
)

type Company struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"size:255;not null"`
    Address     string    `gorm:"size:255"`
    CreatedDate time.Time `gorm:"autoCreateTime"`
}
