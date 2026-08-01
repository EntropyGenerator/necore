package model

import "gorm.io/gorm"

type Glossary struct {
	gorm.Model

	Id      string `gorm:"uniqueIndex;not null" json:"id"`
	Name    string `gorm:"not null" json:"name"`
	Type    string `json:"type"`
	Gallery string `json:"gallery"`
	Content string `json:"content"`
}
