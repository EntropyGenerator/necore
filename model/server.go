package model

import "gorm.io/gorm"

type Server struct {
	gorm.Model

	Id           string `gorm:"uniqueIndex;not null" json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	Realtime     bool   `json:"realtime"`
	OnlineMapUrl string `json:"onlineMapUrl"`
	ServerUrl    string `json:"serverUrl"`
}
