package util

import (
	"necore/config"
	"necore/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tidwall/gjson"
)

func CreateToken(u model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = u.Username
	claims["group"] = u.Group
	claims["department"] = u.Department
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	t, err := token.SignedString([]byte(config.Config("SECRET")))
	return t, err
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

func GetUserDepartmentsFromToken(t *jwt.Token) []string {
	claims := t.Claims.(jwt.MapClaims)["department"]
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
