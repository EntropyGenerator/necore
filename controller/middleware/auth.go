package middleware

import (
	"necore/config"
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthNeeded() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(config.Config("SECRET"))},
		ErrorHandler: jwtError,
		SuccessHandler: func(c *fiber.Ctx) error {
			return validateTokenVersion(c)
		},
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="necore"`)

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Unauthorized",
	})
}

// Authenticate parses the Bearer JWT, loads the user and checks the token
// version without advancing the middleware chain, so it can be used inside a
// regular handler (e.g. guarding the /contents file route). On success the
// user is stored in c.Locals("currentUser") and true is returned; on failure
// the 401 response is written and false is returned.
func Authenticate(c *fiber.Ctx) bool {
	header := c.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(header, "Bearer ") {
		jwtError(c, nil)
		return false
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if tokenString == "" {
		jwtError(c, nil)
		return false
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Config("SECRET")), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		jwtError(c, nil)
		return false
	}

	c.Locals("user", token)
	return resolveCurrentUser(c)
}
