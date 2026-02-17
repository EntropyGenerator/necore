package dao

import (
	"encoding/json"
	"fmt"
	"necore/database"
	"necore/model"
	"os"
	"time"

	"gorm.io/gorm"
)

func getCurrentTime() string {
	currenttime := time.Now()
	newtime := fmt.Sprintf("%d-%s-%d %d:%d:%d", currenttime.Year(), currenttime.Month().String(), currenttime.Day(), currenttime.Hour(), currenttime.Minute(), currenttime.Second())
	return newtime
}

func CreateDocumentNode(parentId string, isFolder bool, private bool, name string, id string, username string) error {
	db := database.GetDocumentDatabase()
	if parentId == id {
		return fmt.Errorf("ParentId and Id cannot be the same")
	}
	contributors, _ := json.Marshal([]string{username})
	node := model.DocumentNode{
		ParentId:     parentId,
		IsFolder:     isFolder,
		Private:      private,
		Name:         name,
		Id:           id,
		Contributors: string(contributors),
		UpdateTime:   getCurrentTime(),
	}
	return db.Create(&node).Error
}

func DeleteDocumentNode(id string) error {
	db := database.GetDocumentDatabase()
	var node model.DocumentNode
	db.Where(&model.DocumentNode{Id: id}).First(&node)

	// Recursively delete all children
	if node.IsFolder {
		var children []model.DocumentNode
		db.Where(&model.DocumentNode{ParentId: id}).Find(&children)
		for _, child := range children {
			DeleteDocumentNode(child.Id)
			db.Where(&model.DocumentNode{Id: child.Id}).Delete(&model.DocumentNode{})
		}
	} else {
		// Delete Files
		os.RemoveAll(fmt.Sprintf("./contents/%s", id))
	}
	return db.Where(&model.DocumentNode{Id: id}).Delete(&model.DocumentNode{}).Error
}

func UpdateDocumentNodeName(id string, name string) error {
	db := database.GetDocumentDatabase()
	return db.Model(&model.DocumentNode{}).
		Where(&model.DocumentNode{Id: id}).
		Updates(model.DocumentNode{
			Name:       name,
			UpdateTime: getCurrentTime()}).Error
}

func UpdateDocumentNodeContent(id string, content string, private bool, username string) error {
	db := database.GetDocumentDatabase()
	var doc model.DocumentNode
	db.Model(&model.DocumentNode{}).Where(&model.DocumentNode{Id: id}).First(&doc)

	// Add username to contributors list
	var contributors []string
	if doc.Contributors != "" {
		json.Unmarshal([]byte(doc.Contributors), &contributors)
	}
	contributors = append(contributors, username)
	deduplicatedContributors := make(map[string]bool, len(contributors))
	for _, contributor := range contributors {
		deduplicatedContributors[contributor] = true
	}
	contributorsList := make([]string, 0)
	for contributor := range deduplicatedContributors {
		if contributor != "" {
			contributorsList = append(contributorsList, contributor)
		}
	}
	newContributors, _ := json.Marshal(contributorsList)

	return db.Model(&model.DocumentNode{}).Where(&model.DocumentNode{Id: id}).
		Updates(model.DocumentNode{
			Content:      content,
			Private:      private,
			Contributors: string(newContributors),
			UpdateTime:   getCurrentTime()}).Error
}

func checkCyclicDocumentNode(parentId string, id string, db *gorm.DB) bool {
	node := model.DocumentNode{}
	db.Where(&model.DocumentNode{Id: parentId}).First(&node)
	if node.Id == id {
		return true
	}
	if node.ParentId == "" {
		return false
	}
	return checkCyclicDocumentNode(node.ParentId, id, db)
}

func UpdateDocumentNodeParentId(id string, parentId string) error {
	db := database.GetDocumentDatabase()
	if parentId == id {
		return fmt.Errorf("ParentId and Id cannot be the same")
	}
	if checkCyclicDocumentNode(parentId, id, db) {
		return fmt.Errorf("Cyclic dependency detected")
	}
	return db.Model(&model.DocumentNode{}).Where(&model.DocumentNode{Id: id}).Updates(model.DocumentNode{ParentId: parentId}).Error
}

func GetDocumentNodeChildren(id string, private bool) ([]model.DocumentNode, error) {
	db := database.GetDocumentDatabase()
	var nodes []model.DocumentNode
	var err error
	if private {
		// all
		err = db.Where(&model.DocumentNode{ParentId: id}).Find(&nodes).Error
	} else {
		// public only
		err = db.Where(map[string]interface{}{"parent_id": id, "private": false}).Find(&nodes).Error
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
		err = db.Where(&model.DocumentNode{Id: id}).First(&node).Error
	} else {
		// public only
		err = db.Where(map[string]interface{}{"id": id, "private": false}).First(&node).Error
	}
	if err != nil {
		return model.DocumentNode{}, err
	}
	return node, nil
}
