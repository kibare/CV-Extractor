package models

import (
    "time"
)

type Candidate struct {
    ID          uint      `gorm:"primaryKey"`
    CVFile      string    `gorm:"size:255"`
    CVFileURL   string    `gorm:"size:255"`
    Name        string    `gorm:"size:255;not null"`
    Email       string    `gorm:"size:255;not null"`
    Domicile    string    `gorm:"size:255"`
    Score       float64   `gorm:"type:float"`
    Skills      string    `gorm:"type:text"`
    IsQualified bool      `gorm:"default:false"`
    PositionID  uint      `gorm:"not null"`
    Position    Position  `gorm:"foreignKey:PositionID"`
    CreatedDate time.Time `gorm:"autoCreateTime"`
}
