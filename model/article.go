package model

import "gorm.io/gorm"

type Article struct {
	gorm.Model

	Id      int    `json:"id"`
	Pin     bool   `json:"pin"`
	Title   string `json:"title"`
	Brief   string `json:"brief"`
	Date    string `json:"date"`
	EndDate string `json:"endDate"`
	Image   string `json:"image"`
	Content string `json:"content"` // json: ["type": "markdown" | "pdf_file" | "image", "content": "string", // markdown content or file url]
}
