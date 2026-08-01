package dao

import (
	"fmt"
	"necore/database"
	"necore/model"
)

func GetWikiTagsByCategory(category string) ([]model.WikiTag, error) {
	db := database.GetWikiDatabase()
	var tags []model.WikiTag
	err := db.Where("category = ?", category).Order("name").Find(&tags).Error
	return tags, err
}

func CreateWikiTag(tag model.WikiTag) error {
	db := database.GetWikiDatabase()
	var existing model.WikiTag
	err := db.Where("category = ? AND name = ?", tag.Category, tag.Name).First(&existing).Error
	if err == nil {
		return fmt.Errorf("tag already exists")
	}
	return db.Create(&tag).Error
}

func DeleteWikiTag(id string) error {
	result := database.GetWikiDatabase().
		Unscoped().
		Where("id = ?", id).
		Delete(&model.WikiTag{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}