package middleware

import (
	"encoding/json"
	"errors"
	"math"
	"necore/database"
	"necore/model"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func validateTokenVersion(c *fiber.Ctx) error {
	if !resolveCurrentUser(c) {
		return nil
	}
	return c.Next()
}

// resolveCurrentUser validates the JWT claims against the stored user and
// token version, then stores the user in c.Locals("currentUser").
// On failure it writes the 401/500 response and returns false.
func resolveCurrentUser(c *fiber.Ctx) bool {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok || token == nil {
		invalidToken(c)
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		invalidToken(c)
		return false
	}

	username, ok := claims["name"].(string)
	if !ok || username == "" {
		invalidToken(c)
		return false
	}

	tokenVersion, ok := getUintClaim(claims, "ver")
	if !ok {
		invalidToken(c)
		return false
	}

	var user model.User
	err := database.GetUserDatabase().
		Select("username", "token_version", "group", "tags").
		Where("username = ?", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		invalidToken(c)
		return false
	}

	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
		return false
	}

	if tokenVersion != user.TokenVersion {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token has been revoked",
		})
		return false
	}

	c.Locals("currentUser", user)
	return true
}

func getUintClaim(claims jwt.MapClaims, key string) (uint, bool) {
	value, ok := claims[key]
	if !ok || value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		if v < 0 || math.Trunc(v) != v {
			return 0, false
		}
		return uint(v), true

	case json.Number:
		parsed, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(parsed), true

	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true

	case uint:
		return v, true

	default:
		return 0, false
	}
}

func invalidToken(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Invalid token",
	})
}
