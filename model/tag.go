package model

import "gorm.io/gorm"

type WikiTag struct {
	gorm.Model

	Id       string `gorm:"uniqueIndex;not null" json:"id"`
	Category string `gorm:"not null;index" json:"category"`
	Name     string `gorm:"not null" json:"name"`
}
