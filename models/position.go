package models

import "time"

type Position struct {
    ID                  uint       `gorm:"primaryKey"`
    Name                string     `gorm:"not null"`
    Education           string     `gorm:"not null"`
    Location            string     `gorm:"not null"`
    MinWorkExp          int        `gorm:"not null"`
    Description         string     `gorm:"not null"`
    Qualification       string     `gorm:"not null"`
    DepartmentID        uint       `gorm:"not null"`
    Department          Department `gorm:"foreignKey:DepartmentID"` // Ensure this relationship is defined
    CreatedDate         time.Time  `gorm:"autoCreateTime"`
    IsResolved          bool       `gorm:"default:false"`
    IsTrash             bool       `gorm:"default:false"`
    IsArchive          bool       `gorm:"default:false"`
    RemovedDate         time.Time  `gorm:"autoUpdateTime"`
    QualifiedCandidates string     `gorm:"type:text"`
    UploadedCV          int        `gorm:"default:0"`
}
