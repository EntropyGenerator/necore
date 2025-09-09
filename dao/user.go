package dao

import (
	"errors"
	"log"
	"necore/config"
	"necore/database"
	"necore/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tidwall/gjson"
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

	gjsonResult := gjson.Get(claims.(string), "@this")
	var result []string
	for _, value := range gjsonResult.Array() {
		result = append(result, value.String())
	}

	return result
}

func GetUserTagsFromToken(t *jwt.Token) []string {
	claims := t.Claims.(jwt.MapClaims)["tags"]
	if claims == nil {
		return []string{}
	}

	gjsonResult := gjson.Get(claims.(string), "@this")
	var result []string
	for _, value := range gjsonResult.Array() {
		result = append(result, value.String())
	}

	return result
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
	db := database.GetInstance()
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
	db := database.GetInstance()
	user := model.User{
		Username: username,
		Password: hash,
	}
	return db.Create(&user).Error
}

func GetAllUsers() ([]model.User, error) {
	db := database.GetInstance()
	var users []model.User
	err := db.Find(&users).Error
	return users, err
}

func DeleteUserByUsername(username string) error {
	db := database.GetInstance()
	return db.Where(&model.User{Username: username}).Delete(&model.User{}).Error
}

func CheckUserPassword(input string, password string) bool {
	return checkPasswordHash(input, password)
}

func UpdateUserPassword(username string, password string) error {
	hash, _ := hashPassword(password)
	db := database.GetInstance()
	return db.Model(&model.User{Username: username}).Update("Password", hash).Error
}

func UpdateUserInfo(username string, group string, tags string) error {
	db := database.GetInstance()
	return db.Model(&model.User{Username: username}).Updates(model.User{Group: group, Tags: tags}).Error
}

func GetUserAvatar(username string) (string, error) {
	db := database.GetInstance()
	var user model.UserAvatar
	if err := db.Where(&model.UserAvatar{Username: username}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	return user.Avatar, nil
}

func UpdateUserAvatar(username string, avatar string) error {
	db := database.GetInstance()
	return db.Save(&model.UserAvatar{Username: username, Avatar: avatar}).Error
}
