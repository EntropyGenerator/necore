package dao

import (
	"encoding/json"
	"errors"
	"log"
	"necore/config"
	"necore/database"
	"necore/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Hash

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func DebugTestPassword() {
	password := "test"
	hash, _ := hashPassword(password)
	log.Println(`Test Password "test":`, hash)
}

// Token

func CreateToken(u model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = u.Username
	claims["group"] = u.Group
	claims["tags"] = u.Tags
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	t, err := token.SignedString([]byte(config.Config("SECRET")))
	return t, err
}

func GetUsernameFromToken(t *jwt.Token) string {
	return t.Claims.(jwt.MapClaims)["username"].(string)
}

func GetUserGroupsFromToken(t *jwt.Token) []string {
	claims := t.Claims.(jwt.MapClaims)["group"]
	if claims == nil {
		return []string{}
	}

	var groups []string
	err := json.Unmarshal([]byte(claims.(string)), &groups)
	if err != nil {
		// log.Println(err)
		return []string{}
	}
	return groups

	// gjsonResult := gjson.Get(claims.(string), "@this")
	// var result []string
	// for _, value := range gjsonResult.Array() {
	// 	result = append(result, value.String())
	// }

	// return result
}

func GetUserTagsFromToken(t *jwt.Token) []string {
	claims := t.Claims.(jwt.MapClaims)["tags"]
	if claims == nil {
		return []string{}
	}

	var tags []string
	err := json.Unmarshal([]byte(claims.(string)), &tags)
	if err != nil {
		// log.Println(err)
		return []string{}
	}
	return tags

	// gjsonResult := gjson.Get(claims.(string), "@this")
	// var result []string
	// for _, value := range gjsonResult.Array() {
	// 	result = append(result, value.String())
	// }

	// return result
}

func IsUserInGroup(t *jwt.Token, group string) bool {
	groups := GetUserGroupsFromToken(t)
	for _, g := range groups {
		if g == group {
			return true
		}
	}
	return false
}

// Database

func GetUserByUsername(u string) (*model.User, error) {
	db := database.GetUserDatabase()
	var user model.User
	if err := db.Where(&model.User{Username: u}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func AddUserByUsername(username string, password string) error {
	hash, _ := hashPassword(password)
	db := database.GetUserDatabase()
	user := model.User{
		Username: username,
		Password: hash,
	}
	return db.Create(&user).Error
}

func GetAllUsers() ([]model.User, error) {
	db := database.GetUserDatabase()
	var users []model.User
	err := db.Find(&users).Error
	return users, err
}

func DeleteUserByUsername(username string) error {
	db := database.GetUserDatabase()
	return db.Delete(&model.User{Username: username}).Error
}

func CheckUserPassword(input string, password string) bool {
	return checkPasswordHash(input, password)
}

func UpdateUserPassword(username string, password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	db := database.GetUserDatabase()
	var user *model.User
	db.Where(&model.User{Username: username}).First(&user)
	return db.Model(&user).Update("Password", hash).Error
}

func UpdateUserInfo(username string, group string, tags string) error {
	db := database.GetUserDatabase()
	var user *model.User
	db.Where(&model.User{Username: username}).First(&user)
	return db.Model(&user).Updates(model.User{Group: group, Tags: tags}).Error
}

func GetUserAvatar(username string) (string, error) {
	db := database.GetUserDatabase()
	var user model.User
	if err := db.Where(&model.User{Username: username}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	return user.Avatar, nil
}

func UpdateUserAvatar(username string, avatar string) error {
	db := database.GetUserDatabase()
	var user *model.User
	db.Where(&model.User{Username: username}).First(&user)
	return db.Model(&user).Updates(model.User{Avatar: avatar}).Error
}
