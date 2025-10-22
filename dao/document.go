package dao

import (
	"encoding/json"
	"fmt"
	"necore/database"
	"necore/model"
	"os"
	"time"
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
		// Delete Files
		os.RemoveAll(fmt.Sprintf("./contents/%s", id))
	}
	return db.Where("id = ?", id).Delete(&model.DocumentNode{}).Error
}

func UpdateDocumentNodeName(id string, name string) error {
	db := database.GetDocumentDatabase()
	return db.Model(&model.DocumentNode{}).Where("id = ?", id).Updates(model.DocumentNode{Name: name, UpdateTime: time.Now().String()}).Error
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
		nodes = []model.DocumentNode{}
		return nodes, err
	}
	return nodes, nil
}

func GetDocumentContent(id string, private bool) (model.DocumentNode, error) {
	db := database.GetDocumentDatabase()
	var node model.DocumentNode
	var err error
	if private {
		// all
		err = db.Where("id = ?", id).First(&node).Error
	} else {
		// public only
		err = db.Where("id = ? and private = ?", id, false).First(&node).Error
	}
	if err != nil {
		return model.DocumentNode{}, err
	}
	return node, nil
}
