package model

import "gorm.io/gorm"

type Server struct {
	gorm.Model

	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	Realtime     bool   `json:"realtime"`
	Online       bool   `json:"online"`
	PlayerCount  int    `json:"playerCount"`
	Capacity     int    `json:"capacity"`
	OnlineMapUrl string `json:"onlineMapUrl"`
	ServerUrl    string `json:"serverUrl"`
}
