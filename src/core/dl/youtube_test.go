package dl

import (
	"testing"
)

func TestBuildYtdlpParamsDoesNotUseCookies(t *testing.T) {
	params := (&youTubeData{}).buildYtdlpParams("abc123", false)

	for i := 0; i < len(params)-1; i++ {
		if params[i] == "--cookies" {
			t.Fatalf("expected yt-dlp params to avoid cookies, got %v", params)
		}
	}

	hasNoConfig := false
	hasNoCookies := false
	for _, arg := range params {
		if arg == "--no-config" {
			hasNoConfig = true
		}
		if arg == "--no-cookies" {
			hasNoCookies = true
		}
	}

	if !hasNoConfig {
		t.Fatalf("expected yt-dlp params to include --no-config, got %v", params)
	}

	if !hasNoCookies {
		t.Fatalf("expected yt-dlp params to include --no-cookies, got %v", params)
	}

	hasNoConfigLocations := false
	hasNoCookiesFromBrowser := false
	hasNoCacheDir := false
	for _, arg := range params {
		if arg == "--no-config-locations" {
			hasNoConfigLocations = true
		}
		if arg == "--no-cookies-from-browser" {
			hasNoCookiesFromBrowser = true
		}
		if arg == "--no-cache-dir" {
			hasNoCacheDir = true
		}
	}

	if !hasNoConfigLocations {
		t.Fatalf("expected yt-dlp params to include --no-config-locations, got %v", params)
	}

	if !hasNoCookiesFromBrowser {
		t.Fatalf("expected yt-dlp params to include --no-cookies-from-browser, got %v", params)
	}

	if !hasNoCacheDir {
		t.Fatalf("expected yt-dlp params to include --no-cache-dir, got %v", params)
	}

	if len(params) == 0 {
		t.Fatal("expected yt-dlp params to be generated")
	}

	hasURL := false
	for _, arg := range params {
		if arg == "https://www.youtube.com/watch?v=abc123" {
			hasURL = true
			break
		}
	}

	if !hasURL {
		t.Fatalf("expected youtube URL to be present in params, got %v", params)
	}
}
