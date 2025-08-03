package handler

import (
	"errors"
	"log"
	"necore/config"
	"necore/database"
	"necore/model"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CheckPasswordHash compare password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func validToken(t *jwt.Token, id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	claims := t.Claims.(jwt.MapClaims)
	uid := int(claims["user_id"].(float64))

	return uid == n
}

func getUserByUsername(u string) (*model.User, error) {
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

func addUser(username string, password string) error {
	db := database.GetInstance()
	user := model.User{
		Username: username,
		Password: password,
	}
	return db.Create(&user).Error
}

// func addUser(username string, password string) {
// }

func Login(c *fiber.Ctx) error {
	// Parse Request Body
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type UserData struct {
		Username   string
		Password   string
		Group      string
		Department string
	}
	input := new(LoginInput)
	log.Println(string(c.Body()))
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error on login request", "err": err})
	}

	// Get User
	var userData UserData
	userModel, err := new(model.User), *new(error)
	userModel, err = getUserByUsername(input.Username)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error", "err": err})
	} else if userModel == nil {
		CheckPasswordHash(input.Password, "")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid identity or password", "err": err})
	} else {
		userData = UserData{
			Username:   userModel.Username,
			Password:   userModel.Password,
			Group:      userModel.Group,
			Department: userModel.Department,
		}
	}

	// Check Password
	if !CheckPasswordHash(input.Password, userData.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid identity or password", "data": nil})
	}

	// Token
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = userData.Username
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(config.Config("SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// DEBUG
	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "data": t})
}

func AddUser(c *fiber.Ctx) error {
	// Parse body
	type NewUser struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	user := new(NewUser)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Review your input", "err": err})
	}

	// Hash Password
	hash, err := hashPassword(user.Password)
	if addUser(user.Username, hash) != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error", "err": err})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success add user", "data": nil})
}
