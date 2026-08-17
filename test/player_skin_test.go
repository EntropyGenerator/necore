package necore_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"necore/service"
)

func fakeSkinStation(t *testing.T, players map[string]string, skinURLs map[string]string) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/yggdrasil/api/users/profiles/minecraft/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/yggdrasil/api/users/profiles/minecraft/")
		uuid, ok := players[name]
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id": uuid, "name": name})
	})
	mux.HandleFunc("/api/yggdrasil/sessionserver/session/minecraft/profile/", func(w http.ResponseWriter, r *http.Request) {
		uuid := strings.TrimPrefix(r.URL.Path, "/api/yggdrasil/sessionserver/session/minecraft/profile/")
		skinURL, ok := skinURLs[uuid]
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		payload := map[string]any{
			"timestamp": 1234567890,
			"textures": map[string]any{
				"SKIN": map[string]string{"url": skinURL},
			},
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		value := base64.StdEncoding.EncodeToString(raw)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   uuid,
			"name": "alice",
			"properties": []map[string]string{
				{"name": "textures", "value": value},
			},
		})
	})
	return httptest.NewServer(mux)
}

func TestPlayerSkin_Resolve(t *testing.T) {
	stationA := fakeSkinStation(t, map[string]string{}, map[string]string{})
	stationB := fakeSkinStation(t, map[string]string{"alice": "uuid-alice-0001"}, map[string]string{"uuid-alice-0001": "https://skin.example/textures/aaa"})
	defer stationA.Close()
	defer stationB.Close()

	t.Setenv("SKIN_STATIONS", stationA.URL+","+stationB.URL)

	url, err := service.ResolvePlayerSkin("alice")
	if err != nil {
		t.Fatalf("resolve alice: %v", err)
	}
	if url != "https://skin.example/textures/aaa" {
		t.Fatalf("skin url = %q", url)
	}

	if _, err := service.ResolvePlayerSkin("nobody"); err == nil {
		t.Fatal("unknown player should fail")
	}

	if _, err := service.ResolvePlayerSkin("  "); err == nil {
		t.Fatal("empty player name should fail")
	}
}

func TestPlayerSkin_PreferFirstStation(t *testing.T) {
	stationA := fakeSkinStation(t, map[string]string{"bob": "uuid-bob"}, map[string]string{"uuid-bob": "https://skin.a/textures/bbb"})
	stationB := fakeSkinStation(t, map[string]string{"bob": "uuid-bob"}, map[string]string{"uuid-bob": "https://skin.b/textures/bbb"})
	defer stationA.Close()
	defer stationB.Close()

	t.Setenv("SKIN_STATIONS", stationA.URL+","+stationB.URL)

	url, err := service.ResolvePlayerSkin("bob")
	if err != nil {
		t.Fatalf("resolve bob: %v", err)
	}
	if !strings.HasPrefix(url, "https://skin.a/") {
		t.Fatalf("expected first station skin, got %q", url)
	}
}
