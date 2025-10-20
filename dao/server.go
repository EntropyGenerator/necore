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
	db.Where(&model.Server{Name: server.Name}).First(&s)
	return db.Model(&s).Updates(server).Error
}

// TODO: fix
func DeleteServer(name string) error {
	db := database.GetServerDatabase()
	return db.Delete(&model.Server{Name: name}).Error
}
