package model

import "gorm.io/gorm"

// type DocumentCategory struct {
// 	gorm.Model

// 	Category string `gorm:"uniqueIndex;not null" json:"category"`
// }

// type DocumentTab struct {
// 	gorm.Model

// 	Tab      string `gorm:"uniqueIndex;not null" json:"tab"`
// 	Category string `gorm:"not null" json:"category"`
// }

type DocumentNode struct {
	gorm.Model

	Id       string `gorm:"uniqueIndex;not null" json:"id"`
	ParentId string `gorm:"not null" json:"parentId"`
	IsFolder bool   `json:"isFolder"`
	Private  bool   `json:"private"`
	Name     string `gorm:"not null" json:"name"`

	Content      string `json:"content"`
	Contributors string `json:"contributors"`
	UpdateTime   string `json:"updateTime"`
}
