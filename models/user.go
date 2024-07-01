package models

import (
    "time"
)

type User struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"size:255;not null"`
    Email       string    `gorm:"size:255;not null;unique"`
    Password    string    `gorm:"size:255;not null"`
    Phone       string    `gorm:"size:255"`
    CompanyID   *uint     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    Company     *Company  `gorm:"foreignKey:CompanyID"`
    CreatedDate time.Time `gorm:"autoCreateTime"`
}
