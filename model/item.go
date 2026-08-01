package model

import "gorm.io/gorm"

type Item struct {
	gorm.Model

	Id       string `gorm:"uniqueIndex;not null" json:"id"`
	Name     string `gorm:"not null" json:"name"`
	Type     string `json:"type"`
	Image    string `json:"image"`
	MaxStack int    `json:"maxStack"`
	Recipe   string `json:"recipe"`
	Content  string `json:"content"`
}
