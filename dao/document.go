package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"necore/database"
	"necore/model"
	"os"
	"time"

	"gorm.io/gorm"
)

func validateDocumentParent(tx *gorm.DB, nodeID string, parentID string) error {
	if parentID == "" {
		return fmt.Errorf("Invalid parent ID")
	}

	if parentID == nodeID {
		return fmt.Errorf("Parent ID cannot be the same as the node ID")
	}

	if parentID == "root" {
		return nil
	}

	var parent model.DocumentNode
	err := tx.Where("id = ?", parentID).First(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("Record not found")
	}
	if err != nil {
		return err
	}

	if !parent.IsFolder {
		return fmt.Errorf("Parent ID must be a folder")
	}

	seen := map[string]struct{}{
		nodeID: {},
	}

	current := parent

	for {
		if _, exists := seen[current.Id]; exists {
			return fmt.Errorf("Circular reference detected")
		}
		seen[current.Id] = struct{}{}

		if current.ParentId == "" || current.ParentId == "root" {
			return nil
		}

		var next model.DocumentNode
		err := tx.Where("id = ?", current.ParentId).First(&next).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("Record not found")
		}
		if err != nil {
			return err
		}

		current = next
	}
}

func getCurrentTime() string {
	currenttime := time.Now()
	newtime := fmt.Sprintf("%d-%s-%d %d:%d:%d", currenttime.Year(), currenttime.Month().String(), currenttime.Day(), currenttime.Hour(), currenttime.Minute(), currenttime.Second())
	return newtime
}

// hasPrivateAncestor reports whether node itself or any of its ancestors is
// marked private. A private node hides its whole subtree: children of a
// private folder must not be exposed through public routes.
func hasPrivateAncestor(db *gorm.DB, node *model.DocumentNode) (bool, error) {
	if node.Private {
		return true, nil
	}
	current := *node
	for current.ParentId != "" && current.ParentId != "root" {
		var parent model.DocumentNode
		err := db.Select("id", "parent_id", "private").
			Where("id = ?", current.ParentId).
			First(&parent).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// broken parent chain: nothing above hides the node
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if parent.Private {
			return true, nil
		}
		current = parent
	}
	return false, nil
}

func CreateDocumentNode(parentId string, isFolder bool, private bool, name string, id string, username string) error {
	db := database.GetDocumentDatabase()

	return db.Transaction(func(tx *gorm.DB) error {
		if err := validateDocumentParent(tx, id, parentId); err != nil {
			return err
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

		return tx.Create(&node).Error
	})
}

func DeleteDocumentNode(id string) error {
	db := database.GetDocumentDatabase()

	var ids []string
	if err := collectDocumentNodeIDs(db, id, &ids); err != nil {
		return err
	}

	if len(ids) == 0 {
		return gorm.ErrRecordNotFound
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Unscoped().
			Where("id IN ?", ids).
			Delete(&model.DocumentNode{})
		if result.Error != nil {
			return result.Error
		}
		return nil
	}); err != nil {
		return err
	}

	for _, nodeID := range ids {
		_ = os.RemoveAll(fmt.Sprintf("./contents/%s", nodeID))
	}

	return nil
}

func collectDocumentNodeIDs(db *gorm.DB, id string, ids *[]string) error {
	var node model.DocumentNode
	err := db.Where("id = ?", id).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}
	if err != nil {
		return err
	}

	*ids = append(*ids, node.Id)

	if node.IsFolder {
		var children []model.DocumentNode
		if err := db.Where("parent_id = ?", node.Id).Find(&children).Error; err != nil {
			return err
		}

		for _, child := range children {
			if err := collectDocumentNodeIDs(db, child.Id, ids); err != nil {
				return err
			}
		}
	}

	return nil
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

	return db.Transaction(func(tx *gorm.DB) error {
		var node model.DocumentNode
		err := tx.Where("id = ?", id).First(&node).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("Record not found")
		}
		if err != nil {
			return err
		}

		if err := validateDocumentParent(tx, id, parentId); err != nil {
			return err
		}

		result := tx.Model(&model.DocumentNode{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"parent_id":   parentId,
				"update_time": getCurrentTime(),
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("Record not found")
		}

		return nil
	})
}

func GetDocumentNodeChildren(id string, private bool) ([]model.DocumentNode, error) {
	db := database.GetDocumentDatabase()
	var nodes []model.DocumentNode
	var err error
	if private {
		// all
		err = db.Where(&model.DocumentNode{ParentId: id}).Find(&nodes).Error
	} else {
		// public only: a private parent (or private ancestor) hides its children
		var parent model.DocumentNode
		perr := db.Select("id", "parent_id", "private").
			Where("id = ?", id).
			First(&parent).Error
		if perr == nil {
			hidden, herr := hasPrivateAncestor(db, &parent)
			if herr != nil {
				return nil, herr
			}
			if hidden {
				return []model.DocumentNode{}, nil
			}
		}
		err = db.Where(map[string]interface{}{"parent_id": id, "private": false}).Find(&nodes).Error
	}
	if err != nil {
		nodes = []model.DocumentNode{}
		return nodes, err
	}
	return nodes, nil
}

// IsDocumentNodeEffectivelyPrivate reports whether the node (or any of its
// ancestors) is marked private. Used by the /contents file guard: files
// uploaded for a private node must not be served anonymously.
func IsDocumentNodeEffectivelyPrivate(id string) (bool, error) {
	db := database.GetDocumentDatabase()
	var node model.DocumentNode
	err := db.Select("id", "parent_id", "private").
		Where("id = ?", id).
		First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// not a document node — treat as public (article/wiki/server files)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return hasPrivateAncestor(db, &node)
}

func GetDocumentContent(id string, private bool) (model.DocumentNode, error) {
	db := database.GetDocumentDatabase()
	var node model.DocumentNode
	var err error
	if private {
		// all
		err = db.Where(&model.DocumentNode{Id: id}).First(&node).Error
	} else {
		// public only: node itself public and no private ancestor
		err = db.Where(map[string]interface{}{"id": id, "private": false}).First(&node).Error
		if err == nil {
			hidden, herr := hasPrivateAncestor(db, &node)
			if herr != nil {
				return model.DocumentNode{}, herr
			}
			if hidden {
				return model.DocumentNode{}, gorm.ErrRecordNotFound
			}
		}
	}
	if err != nil {
		return model.DocumentNode{}, err
	}
	return node, nil
}
