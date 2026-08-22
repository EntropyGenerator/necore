package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"necore/config"

	"github.com/gofiber/fiber/v2"
)

const skinRequestTimeout = 8 * time.Second

var defaultSkinStations = []string{
	"https://skin.nmo.net.cn",
	"https://skin.mualliance.ltd",
}

func skinStations() []string {
	raw := strings.TrimSpace(config.Config("SKIN_STATIONS"))
	if raw == "" {
		return defaultSkinStations
	}

	var stations []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			stations = append(stations, strings.TrimRight(part, "/"))
		}
	}
	if len(stations) == 0 {
		return defaultSkinStations
	}
	return stations
}

func ResolvePlayerAvatar(playerName string) (string, error) {
	playerName = strings.TrimSpace(playerName)
	if playerName == "" {
		return "", errors.New("empty player name")
	}

	client := &http.Client{Timeout: skinRequestTimeout}
	for _, station := range skinStations() {
		avatarURL := fmt.Sprintf(
			"%s/avatar/player/%s",
			station,
			url.PathEscape(playerName),
		)
		resp, err := client.Head(avatarURL)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return avatarURL, nil
		}
	}

	return "", errors.New("avatar not found")
}

func GetPlayerSkin(c *fiber.Ctx) error {
	avatarURL, err := ResolvePlayerAvatar(c.Params("name"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Avatar not found",
		})
	}
	return c.JSON(fiber.Map{
		"skin": avatarURL,
	})
}
