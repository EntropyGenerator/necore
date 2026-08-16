package model

import "gorm.io/gorm"

type Department struct {
	gorm.Model

	Id          string `gorm:"uniqueIndex;not null" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `gorm:"not null;default:0" json:"sortOrder"`
}
