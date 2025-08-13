package handler

import (
	"errors"
	"necore/database"
	"necore/model"
	"necore/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// checkPasswordHash compare password with hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
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

func addUserByUsername(username string, password string) error {
	db := database.GetInstance()
	user := model.User{
		Username: username,
		Password: password,
	}
	return db.Create(&user).Error
}

// Handlers

func Login(c *fiber.Ctx) error {
	// Parse Request Body
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	input := new(LoginInput)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error on login request", "err": err})
	}

	// Get User

	userModel, err := new(model.User), *new(error)
	userModel, err = getUserByUsername(input.Username)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error", "err": err})
	} else if userModel == nil {
		checkPasswordHash(input.Password, "")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid identity or password", "err": err})
	}

	// Check Password
	if !checkPasswordHash(input.Password, userModel.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid identity or password"})
	}

	// Token
	t, err := util.CreateToken(*userModel)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// DEBUG
	type UserInfo struct {
		Username   string `json:"username"`
		Group      string `json:"group"`
		Department string `json:"department"`
	}
	userInfo := UserInfo{
		Username:   userModel.Username,
		Group:      userModel.Group,
		Department: userModel.Department,
	}
	return c.JSON(fiber.Map{
		"token": t,
		"user":  userInfo,
	})
}

func AddUser(c *fiber.Ctx) error {
	// Check if user is admin
	token := c.Locals("user").(*jwt.Token)
	groups := util.GetUserGroupsFromToken(token)
	var isAdmin bool = false
	for _, g := range groups {
		if g == "admin" {
			isAdmin = true
		}
	}
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Parse body
	type NewUser struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	user := new(NewUser)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Review your input", "err": err})
	}

	// Hash Password
	hash, err := hashPassword(user.Password)
	if addUserByUsername(user.Username, hash) != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error", "err": err})
	}

	return c.JSON(fiber.Map{})
}
