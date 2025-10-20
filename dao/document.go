package dao

import (
	"encoding/json"
	"necore/database"
	"necore/model"
)

// func CreateDocumentCategory(categoryName string) error {
// 	db := database.GetDocumentDatabase()
// 	category := model.DocumentCategory{
// 		Category: categoryName,
// 	}
// 	return db.Create(&category).Error
// }

func CreateDocumentNode(parentId string, isFolder bool, private bool, name string, id string) error {
	db := database.GetDocumentDatabase()
	node := model.DocumentNode{
		ParentId: parentId,
		IsFolder: isFolder,
		Private:  private,
		Name:     name,
		Id:       id,
	}
	return db.Create(&node).Error
}

func DeleteDocumentNode(id string) error {
	db := database.GetDocumentDatabase()
	var node model.DocumentNode
	db.Where("id = ?", id).First(&node)

	// Recursively delete all children
	if node.IsFolder {
		var children []model.DocumentNode
		db.Where("parentId = ?", id).Find(&children)
		for _, child := range children {
			DeleteDocumentNode(child.Id)
			db.Where("id = ?", child.Id).Delete(&model.DocumentNode{})
		}
	} else {
		// TODO: delete file
	}
	return db.Where("id = ?", id).Delete(&model.DocumentNode{}).Error
}

func UpdateDocumentNodeName(id string, name string) error {
	db := database.GetDocumentDatabase()
	return db.Model(&model.DocumentNode{}).Where("id = ?", id).Update("name", name).Error
}

func UpdateDocumentNodeContent(id string, content string, private bool, username string) error {
	db := database.GetDocumentDatabase()
	var doc model.DocumentNode
	db.Model(&model.DocumentNode{}).Where("id = ?", id).First(&doc)

	// Add username to contributors list
	var contributors []string
	if doc.Contributors != "" {
		json.Unmarshal([]byte(doc.Contributors), &contributors)
	}
	deduplicatedContributors := make(map[string]bool, len(contributors))
	for _, contributor := range contributors {
		deduplicatedContributors[contributor] = true
	}
	contributorsList := make([]string, 0, len(deduplicatedContributors))
	for contributor := range deduplicatedContributors {
		contributorsList = append(contributorsList, contributor)
	}
	newContributors, _ := json.Marshal(contributorsList)

	return db.Model(&model.DocumentNode{}).Where("id = ?", id).Updates(model.DocumentNode{Content: content, Private: private, Contributors: string(newContributors)}).Error
}

func UpdateDocumentNodeParentId(id string, parentId string) error {
	db := database.GetDocumentDatabase()
	return db.Model(&model.DocumentNode{}).Where("id = ?", id).Update("parentId", parentId).Error
}

// func GetDocumentNode(id string) (model.DocumentNode, error) {
// 	db := database.GetDocumentDatabase()
// 	var node model.DocumentNode
// 	err := db.Where("id = ?", id).First(&node).Error
// 	if err != nil {
// 		return model.DocumentNode{}, err
// 	}
// 	return node, nil
// }

func GetDocumentNodeChildren(id string, private bool) ([]model.DocumentNode, error) {
	db := database.GetDocumentDatabase()
	var nodes []model.DocumentNode
	var err error
	if private {
		// all
		err = db.Where("parentId = ?", id).Find(&nodes).Error
	} else {
		// public only
		err = db.Where("parentId = ? and private = ?", id, false).Find(&nodes).Error
	}
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// func GetDocumentAllCategory() ([]model.DocumentCategory, error) {
// 	db := database.GetDocumentDatabase()
// 	var categories []model.DocumentCategory
// 	err := db.Order("created_at desc").Find(&categories).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return categories, nil
// }

// func DeleteDocumentCategory(categoryName string) error {
// 	db := database.GetDocumentDatabase()
// 	db.Where("category = ?", categoryName).Delete(&model.Document{})
// 	db.Where("category = ?", categoryName).Delete(&model.DocumentTab{})
// 	return db.Delete(&model.DocumentCategory{Category: categoryName}).Error
// }

// func CreateDocumentTab(categoryName string, tabName string) error {
// 	db := database.GetDocumentDatabase()
// 	tab := model.DocumentTab{
// 		Category: categoryName,
// 		Tab:      tabName,
// 	}
// 	return db.Create(&tab).Error
// }

// func DeleteDocumentTab(categoryName string, tabName string) error {
// 	db := database.GetDocumentDatabase()
// 	db.Where("category = ? and tab = ?").Delete(&model.Document{})
// 	return db.Delete(&model.DocumentTab{Category: categoryName, Tab: tabName}).Error
// }

// func GetDocumentAllTab(categoryName string) ([]model.DocumentTab, error) {
// 	db := database.GetDocumentDatabase()
// 	var tabs []model.DocumentTab
// 	err := db.Where("category = ?", categoryName).Order("created_at desc").Find(&tabs).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tabs, nil
// }

// func CreateDocument(id string) error {
// 	db := database.GetDocumentDatabase()
// 	document := model.Document{
// 		Id: id,
// 	}
// 	return db.Create(&document).Error
// }

// func UpdateDocument(doc model.Document) error {
// 	db := database.GetDocumentDatabase()
// 	return db.Save(&doc).Error
// }

// func DeleteDocument(id string) error {
// 	db := database.GetDocumentDatabase()
// 	return db.Delete(&model.Document{Id: id}).Error
// }

// func GetDocument(id string, isAdmin bool) (model.Document, error) {
// 	db := database.GetDocumentDatabase()
// 	var document model.Document
// 	var err error
// 	if isAdmin {
// 		err = db.Where("id = ?", id).First(&document).Error
// 	} else {
// 		err = db.Where("id = ? and private = ?", id, false).First(&document).Error
// 	}
// 	if err != nil {
// 		return document, err
// 	}
// 	return document, nil
// }

// func GetDocumentListByNum(start int, limit int, isAdmin bool) ([]model.Document, error) {
// 	db := database.GetDocumentDatabase()
// 	var documents []model.Document
// 	var err error
// 	if isAdmin {
// 		err = db.Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
// 	} else {
// 		err = db.Where("private = ?", false).Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return documents, nil
// }

// func GetDocumentListByClass(category string, tab string, isAdmin bool) ([]model.Document, error) {
// 	db := database.GetDocumentDatabase()
// 	var documents []model.Document
// 	var err error
// 	if isAdmin {
// 		err = db.Where("category = ? and tab = ?", category, tab).Order("created_at desc").Find(&documents).Error
// 	} else {
// 		err = db.Where("category = ? and tab = ? and private = ?", category, tab, false).Order("created_at desc").Find(&documents).Error
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return documents, nil
// }

// func SearchDocument(keyword string, start int, limit int, isAdmin bool) ([]model.Document, error) {
// 	db := database.GetDocumentDatabase()
// 	var documents []model.Document
// 	var err error
// 	if isAdmin {
// 		err = db.Where("title like ?", "%"+keyword+"%").Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
// 	} else {
// 		err = db.Where("title like ? and private = ?", "%"+keyword+"%", false).Order("created_at desc").Offset(start).Limit(limit).Find(&documents).Error
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return documents, nil
// }
