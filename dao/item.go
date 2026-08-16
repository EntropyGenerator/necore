package dao

import (
	"fmt"
	"necore/database"
	"necore/model"
)

func GetItemList() ([]model.Item, error) {
	db := database.GetWikiDatabase()
	var list []model.Item
	err := db.Find(&list).Error
	return list, err
}

func GetItemById(id string) (model.Item, error) {
	db := database.GetWikiDatabase()
	var entry model.Item
	err := db.Where("id = ?", id).First(&entry).Error
	return entry, err
}

func CreateItem(entry model.Item) error {
	db := database.GetWikiDatabase()
	return db.Create(&entry).Error
}

func UpdateItem(entry model.Item) error {
	result := database.GetWikiDatabase().
		Model(&model.Item{}).
		Where("id = ?", entry.Id).
		Updates(map[string]any{
			"name":      entry.Name,
			"type":      entry.Type,
			"image":     entry.Image,
			"max_stack": entry.MaxStack,
			"recipe":    entry.Recipe,
			"content":   entry.Content,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}

func DeleteItem(id string) error {
	result := database.GetWikiDatabase().
		Unscoped().
		Where("id = ?", id).
		Delete(&model.Item{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}
