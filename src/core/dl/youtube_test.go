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

	if len(params) == 0 {
		t.Fatal("expected yt-dlp params to be generated")
	}

	if params[len(params)-2] != "https://www.youtube.com/watch?v=abc123" {
		t.Fatalf("expected youtube URL at the end of params, got %v", params)
	}
}
