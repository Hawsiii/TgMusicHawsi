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

    if len(params) < 3 || params[len(params)-3] != "https://www.youtube.com/watch?v=abc123" {
        t.Fatalf("expected youtube URL near the end of params, got %v", params)
    }
}
