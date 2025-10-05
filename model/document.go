package model

import "gorm.io/gorm"

type DocumentCategory struct {
	gorm.Model

	Category string `gorm:"uniqueIndex;not null" json:"category"`
}

type DocumentTab struct {
	gorm.Model

	Tab      string `gorm:"uniqueIndex;not null" json:"tab"`
	Category string `gorm:"not null" json:"category"`
}

type Document struct {
	gorm.Model

	Id           string `gorm:"uniqueIndex;not null" json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Content      string `json:"content"`
	Tab          string `json:"tab"`
	Category     string `json:"category"`
	Contributors string `json:"contributors"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
	Private      bool   `json:"private"`
	Priority     int    `json:"priority"`
}
