package models

import (
    "time"
)

type Department struct {
    ID          uint       `gorm:"primaryKey"`
    Name        string     `gorm:"size:255;not null"`
    CompanyID   uint       `gorm:"not null"`
    Company     Company    `gorm:"foreignKey:CompanyID"`
    Positions   []Position `gorm:"foreignKey:DepartmentID"`
    CreatedDate time.Time  `gorm:"autoCreateTime"`
}
