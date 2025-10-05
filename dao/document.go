package dao

import (
	"necore/database"
	"necore/model"
)

func CreateDocumentCategory(categoryName string) error {
	db := database.GetDocumentDatabase()
	category := model.DocumentCategory{
		Category: categoryName,
	}
	return db.Create(&category).Error
}

func GetDocumentAllCategory() ([]model.DocumentCategory, error) {
	db := database.GetDocumentDatabase()
	var categories []model.DocumentCategory
	err := db.Order("created_at desc").Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func DeleteDocumentCategory(categoryName string) error {
	db := database.GetDocumentDatabase()
	db.Where("category = ?", categoryName).Delete(&model.Document{})
	db.Where("category = ?", categoryName).Delete(&model.DocumentTab{})
	return db.Delete(&model.DocumentCategory{Category: categoryName}).Error
}

func CreateDocumentTab(categoryName string, tabName string) error {
	db := database.GetDocumentDatabase()
	tab := model.DocumentTab{
		Category: categoryName,
		Tab:      tabName,
	}
	return db.Create(&tab).Error
}

func DeleteDocumentTab(categoryName string, tabName string) error {
	db := database.GetDocumentDatabase()
	db.Where("category = ? and tab = ?").Delete(&model.Document{})
	return db.Delete(&model.DocumentTab{Category: categoryName, Tab: tabName}).Error
}

func GetDocumentAllTab(categoryName string) ([]model.DocumentTab, error) {
	db := database.GetDocumentDatabase()
	var tabs []model.DocumentTab
	err := db.Where("category = ?", categoryName).Order("created_at desc").Find(&tabs).Error
	if err != nil {
		return nil, err
	}
	return tabs, nil
}

func CreateDocument(id string) error {
	db := database.GetDocumentDatabase()
	document := model.Document{
		Id: id,
	}
	return db.Create(&document).Error
}

func UpdateDocument(doc model.Document) error {
	db := database.GetDocumentDatabase()
	return db.Save(&doc).Error
}

func DeleteDocument(id string) error {
	db := database.GetDocumentDatabase()
	return db.Delete(&model.Document{Id: id}).Error
}

func GetDocument(id string, isAdmin bool) (model.Document, error) {
	db := database.GetDocumentDatabase()
	var document model.Document
	var err error
	if isAdmin {
		err = db.Where("id = ?", id).First(&document).Error
	} else {
		err = db.Where("id = ? and private = ?", id, false).First(&document).Error
	}
	if err != nil {
		return document, err
	}
	return document, nil
}

func GetDocumentListByNum(start int, limit int, isAdmin bool) ([]model.Document, error) {
	db := database.GetDocumentDatabase()
	var documents []model.Document
	var err error
	if isAdmin {
		err = db.Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
	} else {
		err = db.Where("private = ?", false).Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
	}
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func GetDocumentListByClass(category string, tab string, isAdmin bool) ([]model.Document, error) {
	db := database.GetDocumentDatabase()
	var documents []model.Document
	var err error
	if isAdmin {
		err = db.Where("category = ? and tab = ?", category, tab).Order("created_at desc").Find(&documents).Error
	} else {
		err = db.Where("category = ? and tab = ? and private = ?", category, tab, false).Order("created_at desc").Find(&documents).Error
	}
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func SearchDocument(keyword string, start int, limit int, isAdmin bool) ([]model.Document, error) {
	db := database.GetDocumentDatabase()
	var documents []model.Document
	var err error
	if isAdmin {
		err = db.Where("title like ?", "%"+keyword+"%").Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
	} else {
		err = db.Where("title like ? and private = ?", "%"+keyword+"%", false).Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
	}
	if err != nil {
		return nil, err
	}
	return documents, nil
}
