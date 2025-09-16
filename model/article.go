package model

import "gorm.io/gorm"

//"information" | "magazine" | "notice" | "activity" | "document"

type Article struct {
	gorm.Model

	Id       string `gorm:"uniqueIndex;not null" json:"id"`
	Pin      bool   `json:"pin"`
	Title    string `json:"title"`
	Brief    string `json:"brief"`
	Date     string `json:"date"`
	EndDate  string `json:"endDate"`
	Image    string `json:"image"`
	Content  string `json:"content"` // json: ["type": "markdown" | "pdf_file" | "image", "content": "string", // markdown content or file url]
	Author   string `json:"author"`
	Category string `json:"category"`
}
