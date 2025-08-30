package entities

import (
	"gorm.io/gorm"
)

type Store struct {
    gorm.Model
    IDUser   uint    `gorm:"not null"`
    NamaToko *string `gorm:"size:255;default:null"`
    UrlFoto  *string `gorm:"size:255;default:null"`
}


func (Store) TableName() string {
	return "Toko"
}