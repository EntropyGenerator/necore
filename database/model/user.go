package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Username   string `gorm:"uniqueIndex;not null" json:"username"`
	Password   string `gorm:"not null" json:"password"` // sha256 hashed
	Group      string `json:"group"`                    // json array: []string
	Department string `json:"department"`               // json array: []string
}
