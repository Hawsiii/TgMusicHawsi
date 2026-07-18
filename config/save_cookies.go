/*
 * TgMusicBot - Telegram Music Bot
 * Copyright (c) 2025-2026 Ashok Shau
 */

package config

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const cookiesDr = "src/cookies"

var cookieMu sync.RWMutex

// fetchContent downloads cookie content.
func fetchContent(url string) (string, error) {
	parts := strings.Split(strings.Trim(url, "/"), "/")
	id := parts[len(parts)-1]

	rawURL := url
	if strings.Contains(url, "pastebin.com") {
		rawURL = fmt.Sprintf("https://pastebin.com/raw/%s", id)
	} else if strings.Contains(url, "batbin.me") {
		rawURL = fmt.Sprintf("https://batbin.me/raw/%s", id)
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// saveContent writes one cookie file.
func saveContent(url, content string) (string, error) {
	parts := strings.Split(strings.Trim(url, "/"), "/")
	filename := parts[len(parts)-1]

	if filename == "" {
		filename = "cookie"
	}

	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	path := filepath.Join(cookiesDr, filename)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

// saveAllCookies downloads every cookie and refreshes CookiesPath safely.
func saveAllCookies(urls []string) {
	var paths []string

	for _, url := range urls {
		content, err := fetchContent(url)
		if err != nil {
			slog.Warn("Cookie download failed", "url", url, "error", err)
			continue
		}

		path, err := saveContent(url, content)
		if err != nil {
			slog.Warn("Cookie save failed", "url", url, "error", err)
			continue
		}

		paths = append(paths, path)
		slog.Info("Cookie loaded", "file", path)
	}

	cookieMu.Lock()
	CookiesPath = paths
	cookieMu.Unlock()

	slog.Info("Cookie pool ready", "count", len(CookiesPath))
}

// GetCookiePaths returns a thread-safe copy of all cookie paths.
func GetCookiePaths() []string {
	cookieMu.RLock()
	defer cookieMu.RUnlock()

	out := make([]string, len(CookiesPath))
	copy(out, CookiesPath)

	return out
}

