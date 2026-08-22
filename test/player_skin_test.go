package necore_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"necore/service"
)

func fakeSkinStation(t *testing.T, players map[string]bool) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/avatar/player/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/avatar/player/")
		if !players[name] {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(mux)
}

func TestPlayerAvatar_Resolve(t *testing.T) {
	stationA := fakeSkinStation(t, map[string]bool{})
	stationB := fakeSkinStation(t, map[string]bool{"alice": true})
	defer stationA.Close()
	defer stationB.Close()

	t.Setenv("SKIN_STATIONS", stationA.URL+","+stationB.URL)

	url, err := service.ResolvePlayerAvatar("alice")
	if err != nil {
		t.Fatalf("resolve alice: %v", err)
	}
	if url != stationB.URL+"/avatar/player/alice" {
		t.Fatalf("avatar url = %q", url)
	}

	if _, err := service.ResolvePlayerAvatar("nobody"); err == nil {
		t.Fatal("unknown player should fail")
	}

	if _, err := service.ResolvePlayerAvatar("  "); err == nil {
		t.Fatal("empty player name should fail")
	}
}

func TestPlayerAvatar_PreferFirstStation(t *testing.T) {
	stationA := fakeSkinStation(t, map[string]bool{"bob": true})
	stationB := fakeSkinStation(t, map[string]bool{"bob": true})
	defer stationA.Close()
	defer stationB.Close()

	t.Setenv("SKIN_STATIONS", stationA.URL+","+stationB.URL)

	url, err := service.ResolvePlayerAvatar("bob")
	if err != nil {
		t.Fatalf("resolve bob: %v", err)
	}
	if !strings.HasPrefix(url, stationA.URL) {
		t.Fatalf("expected first station avatar, got %q", url)
	}
}
