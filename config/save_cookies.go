/*
 * TgMusicBot - Telegram Music Bot
 * Copyright (c) 2025-2026 Ashok Shau
 */

package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const cookiesDr = "src/cookies"

func fetchContent(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("empty cookie source")
	}

	if strings.Contains(url, "github.com") {
		return fetchFromGitHub(url)
	}

	parts := strings.Split(strings.Trim(url, "/"), "/")
	id := parts[len(parts)-1]

	rawURL := url
	if strings.Contains(url, "pastebin.com") {
		rawURL = fmt.Sprintf("https://pastebin.com/raw/%s", id)
	} else if strings.Contains(url, "batbin.me") {
		rawURL = fmt.Sprintf("https://batbin.me/raw/%s", id)
	}

	return fetchURL(rawURL, "")
}

func fetchURL(rawURL, githubToken string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}

	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.raw+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func fetchFromGitHub(url string) (string, error) {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN is not set")
	}

	repo, path, ref, err := parseGitHubURL(url)
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, path)
	if ref != "" {
		apiURL += "?ref=" + ref
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("github http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	contentType := resp.Header.Get("Content-Type")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// If GitHub returns JSON metadata instead of raw file bytes, decode it.
	if strings.Contains(contentType, "application/json") {
		type githubContent struct {
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}
		var gc githubContent
		if err := jsonUnmarshal(body, &gc); err != nil {
			return "", err
		}
		if gc.Encoding == "base64" {
			clean := strings.ReplaceAll(gc.Content, "\n", "")
			decoded, err := base64.StdEncoding.DecodeString(clean)
			if err != nil {
				return "", err
			}
			return string(decoded), nil
		}
		return gc.Content, nil
	}

	return string(body), nil
}

func parseGitHubURL(url string) (repo string, path string, ref string, err error) {
	// Supports:
	// https://github.com/OWNER/REPO
	// https://github.com/OWNER/REPO/blob/BRANCH/PATH
	// https://raw.githubusercontent.com/OWNER/REPO/BRANCH/PATH
	u := strings.TrimSpace(url)

	if strings.Contains(u, "raw.githubusercontent.com") {
		parts := strings.Split(strings.Trim(u, "/"), "/")
		if len(parts) < 5 {
			return "", "", "", fmt.Errorf("invalid raw github url")
		}
		// .../OWNER/REPO/BRANCH/PATH...
		repo = parts[len(parts)-4] + "/" + parts[len(parts)-3]
		ref = parts[len(parts)-2]
		path = strings.Join(parts[len(parts)-1:], "/")
		return repo, path, ref, nil
	}

	if !strings.Contains(u, "github.com/") {
		return "", "", "", fmt.Errorf("not a github url")
	}

	u = strings.TrimPrefix(u, "https://github.com/")
	u = strings.TrimPrefix(u, "http://github.com/")
	u = strings.Trim(u, "/")

	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid github url")
	}

	repo = parts[0] + "/" + parts[1]

	// repo root
	if len(parts) == 2 {
		return repo, "", "", nil
	}

	// blob/branch/path/to/file
	if len(parts) >= 5 && parts[2] == "blob" {
		ref = parts[3]
		path = strings.Join(parts[4:], "/")
		return repo, path, ref, nil
	}

	// fallback: treat trailing path after repo as path
	if len(parts) > 2 {
		path = strings.Join(parts[2:], "/")
	}

	return repo, path, "", nil
}

func jsonUnmarshal(data []byte, v any) error {
	type jsonUnmarshaler interface {
		Unmarshal([]byte, any) error
	}
	return json.Unmarshal(data, v)
}

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

func saveAllCookies(urls []string) {
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

		slog.Info("Cookie loaded", "file", path)
	}
}
