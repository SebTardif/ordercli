package cli

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/steipete/ordercli/internal/config"
)

func TestAuthedRefreshKeepsTokenAndSaves(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","expires_in":3600}`))
	}))
	t.Cleanup(srv.Close)

	path := filepath.Join(t.TempDir(), "config.json")
	stored := &config.FoodoraConfig{
		BaseURL:       srv.URL + "/",
		AccessToken:   "old-access",
		RefreshToken:  "old-refresh",
		ExpiresAt:     time.Now().Add(-time.Hour),
		OAuthClientID: "android",
		ClientSecret:  "sec",
	}
	if err := config.Save(path, config.Config{Version: 1, Providers: config.Providers{Foodora: stored}}); err != nil {
		t.Fatal(err)
	}
	st := &state{}
	st.configPath = path
	if err := st.load(); err != nil {
		t.Fatal(err)
	}
	if _, err := newAuthedClient(st); err != nil {
		t.Fatal(err)
	}
	if got := st.foodora().RefreshToken; got != "old-refresh" {
		t.Fatalf("refresh token = %q; want old-refresh", got)
	}
	if got := st.foodora().AccessToken; got != "new-access" {
		t.Fatalf("access token = %q; want new-access", got)
	}

	reloaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Foodora()
	if got.AccessToken != "new-access" || got.RefreshToken != "old-refresh" {
		t.Fatalf("saved access=%q refresh=%q; want new-access and old-refresh", got.AccessToken, got.RefreshToken)
	}
}
