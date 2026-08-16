package service

import (
	"encoding/base64"
	"encoding/json"
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

// skinStations 返回皮肤站列表（Blessing Skin / Yggdrasil），
// 可用 .env 的 SKIN_STATIONS（逗号分隔）覆盖。
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

type yggdrasilProfile struct {
	Properties []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"properties"`
}

// yggdrasilUuidByName 通过皮肤站 Yggdrasil 的姓名查找接口获取玩家 UUID。
func yggdrasilUuidByName(station, playerName string) (string, error) {
	client := &http.Client{Timeout: skinRequestTimeout}
	endpoint := fmt.Sprintf(
		"%s/api/yggdrasil/api/users/profiles/minecraft/%s",
		station,
		url.PathEscape(playerName),
	)
	resp, err := client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("player not found")
	}

	var data struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Id, nil
}

// yggdrasilSkinUrl 通过 UUID 获取皮肤纹理 URL（解码 textures 属性）。
func yggdrasilSkinUrl(station, uuid string) (string, error) {
	client := &http.Client{Timeout: skinRequestTimeout}
	endpoint := fmt.Sprintf(
		"%s/api/yggdrasil/sessionserver/session/minecraft/profile/%s",
		station,
		uuid,
	)
	resp, err := client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("profile not found")
	}

	var profile yggdrasilProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", err
	}

	for _, property := range profile.Properties {
		if property.Name != "textures" {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(property.Value)
		if err != nil {
			continue
		}
		var textures struct {
			Textures struct {
				Skin struct {
					Url string `json:"url"`
				} `json:"SKIN"`
			} `json:"textures"`
		}
		if err := json.Unmarshal(raw, &textures); err != nil {
			continue
		}
		if textures.Textures.Skin.Url != "" {
			return textures.Textures.Skin.Url, nil
		}
	}

	return "", errors.New("no skin texture")
}

// ResolvePlayerSkin 依次尝试各皮肤站，返回玩家的皮肤纹理 URL。
func ResolvePlayerSkin(playerName string) (string, error) {
	playerName = strings.TrimSpace(playerName)
	if playerName == "" {
		return "", errors.New("empty player name")
	}

	for _, station := range skinStations() {
		stationUUID, err := yggdrasilUuidByName(station, playerName)
		if err != nil || stationUUID == "" {
			continue
		}
		skinURL, err := yggdrasilSkinUrl(station, stationUUID)
		if err != nil || skinURL == "" {
			continue
		}
		return skinURL, nil
	}

	return "", errors.New("skin not found")
}

// GetPlayerSkin 处理器：GET /server/skin/:name
func GetPlayerSkin(c *fiber.Ctx) error {
	skinURL, err := ResolvePlayerSkin(c.Params("name"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Skin not found",
		})
	}
	return c.JSON(fiber.Map{
		"skin": skinURL,
	})
}
