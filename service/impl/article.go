package impl

import (
	"necore/database"
	"necore/model"
)

// Database

func CreateArticle(id string) error {
	db := database.GetInstance()
	article := model.Article{
		Id: id,
	}
	return db.Create(&article).Error
}

func UpdateArticle(updatedArticle model.Article) error {
	db := database.GetInstance()
	return db.Save(&updatedArticle).Error
}

func GetArticle(id string) (*model.Article, error) {
	db := database.GetInstance()
	article := model.Article{}
	if err := db.Where(&model.Article{Id: id}).First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func GetArticleCountByCategory(category string) (int64, error) {
	db := database.GetInstance()
	var count int64
	if err := db.Model(&model.Article{}).Where("category = ?", category).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetArticleList(target string, page int, pageSize int, pin bool) ([]model.Article, error) {
	db := database.GetInstance()
	var articles []model.Article
	err := db.Where("target = ? AND pin = ?", target, pin).
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&articles).Error
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func DeleteArticle(id string) error {
	db := database.GetInstance()
	return db.Delete(&model.Article{Id: id}).Error
}
