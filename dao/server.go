package dao

import (
	"necore/database"
	"necore/model"
)

func GetServerList() ([]model.Server, error) {
	db := database.GetServerDatabase()
	var servers []model.Server
	err := db.Find(&servers).Error
	return servers, err
}

func AddServer(server model.Server) error {
	db := database.GetServerDatabase()
	return db.Create(&server).Error
}

func UpdateServer(server model.Server) error {
	db := database.GetServerDatabase()
	var s *model.Server
	db.Where(&model.Server{Id: server.Id}).First(&s)
	return db.Model(&s).Updates(server).Error
}

func DeleteServer(id string) error {
	db := database.GetServerDatabase()
	return db.Where(&model.Server{Id: id}).Delete(&model.Server{}).Error
}
