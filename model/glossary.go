package model

import "gorm.io/gorm"

var GlossaryTypes = []string{"服务器", "社群", "概念", "地理", "人物", "其它"}

type Glossary struct {
	gorm.Model

	Id      string `gorm:"uniqueIndex;not null" json:"id"`
	Name    string `gorm:"not null" json:"name"`
	Type    string `json:"type"`
	Gallery string `json:"gallery"`
	Content string `json:"content"`
}
