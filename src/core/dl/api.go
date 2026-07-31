/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package dl

import (
	"log/slog"

	"ashokshau/tgmusic/config"
	"ashokshau/tgmusic/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// apiData provides a unified interface for fetching track and playlist information from various music platforms via an API gateway.
type apiData struct {
	Query    string
	ApiUrl   string
	APIKey   string
	Patterns map[string]*regexp.Regexp
}

var apiPatterns = map[string]*regexp.Regexp{
	utils.Apple:      regexp.MustCompile(`(?i)^https?:\/\/music\.apple\.com\/[a-zA-Z-]+\/(?:song\/(?:[^\/]+\/)?\d+|album\/[^\/]+\/\d+(?:\?i=\d+)?|playlist\/[^\/]+\/pl\.[\w.-]+|artist\/[^\/]+\/\d+)(?:\?.*)?$`),
	utils.Spotify:    regexp.MustCompile(`(?i)^(https?://)?([a-z0-9-]+\.)*spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+(\?.*)?$`),
	utils.YouTube:    regexp.MustCompile(`(?i)^(?:https?://)?(?:www\.)?(?:youtube\.com|youtu\.be)/.*$`),
	utils.JioSaavn:   regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?jiosaavn\.com\/(song|album|playlist|featured)\/[^\/]+\/([A-Za-z0-9_]+)`),
	utils.Deezer:     regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?deezer\.com\/(?:[a-z]{2}\/)?(track|album|playlist)\/(\d+)`),
	utils.SoundCloud: regexp.MustCompile(`(?i)^(https?://)?(www\.)?soundcloud\.com/[a-zA-Z0-9_-]+/(sets/)?[a-zA-Z0-9._-]+(\?.*)?$`),
	utils.Gaana:      regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?gaana\.com\/(song|album|playlist|artist)\/([A-Za-z0-9\-]+)`),
	utils.Tidal:      regexp.MustCompile(`(?i)https?:\/\/(?:www\.|listen\.)?tidal\.com\/(?:browse\/)?(track|album|playlist)\/([a-zA-Z0-9-]+)(?:[\/?].*)?`),
	utils.MXPlayer:   regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?mxplayer\.in\/(?:show|movie)\/.*`),
	utils.Twitch:     regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?twitch\.tv\/(?:videos|[\w._-]+\/video)\/\d+`),
	utils.TwitchClip: regexp.MustCompile(
		`(?i)https?:\/\/(?:www\.|m\.)?(?:` +
			`twitch\.tv\/clip\/[\w-]+|` +
			`clips\.twitch\.tv\/[\w-]+|` +
			`twitch\.tv\/[\w-]+\/clip\/[\w-]+` +
			`)`,
	),
	utils.Kick:     regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/videos\/[a-fA-F0-9-]+`),
	utils.KickClip: regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/clips\/[\w-]+`),
}

// newApiData creates and initializes a new apiData instance with the provided query.
func newApiData(query string) *apiData {
	return &apiData{
		Query:    strings.TrimSpace(query),
		ApiUrl:   strings.TrimRight(config.ApiUrl, "/"),
		APIKey:   config.ApiKey,
		Patterns: apiPatterns,
	}
}

func (a *apiData) isValid() bool {
	if a.Query == "" || a.ApiUrl == "" || a.APIKey == "" {
		return false
	}

	for _, pattern := range a.Patterns {
		if pattern.MatchString(a.Query) {
			return true
		}
	}
	return false
}

// getInfo retrieves metadata for a track or playlist from the API.
func (a *apiData) getInfo() (utils.PlatformTracks, error) {
	if !a.isValid() {
		return utils.PlatformTracks{}, errors.New("the provided URL is invalid or the platform is not supported")
	}

	params := url.Values{"url": {a.Query}}
	resp, err := a.apiRequest("/api/get_url", params)
	if err != nil {
		return utils.PlatformTracks{}, fmt.Errorf("the GetInfo request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return utils.PlatformTracks{}, fmt.Errorf("unexpected status code while fetching info: %s", resp.Status)
	}

	var data utils.PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return utils.PlatformTracks{}, fmt.Errorf("failed to decode the GetInfo response: %w", err)
	}
	return data, nil
}

// search queries the API for a track.
func (a *apiData) search() (utils.PlatformTracks, error) {
	if a.isValid() {
		return a.getInfo()
	}

	params := url.Values{
		"query": {a.Query},
		"limit": {"5"},
	}
	resp, err := a.apiRequest("/api/search", params)
	if err != nil {
		return utils.PlatformTracks{}, fmt.Errorf("the search request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return utils.PlatformTracks{}, fmt.Errorf("unexpected status code during search: %s", resp.Status)
	}

	var data utils.PlatformTracks
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		slog.Warn("Failed to decode search response", "error", err)
		return utils.PlatformTracks{}, fmt.Errorf("failed to decode the search response: %w", err)
	}

	return data, nil
}

// getTrack retrieves detailed information for a single track from the API.
func (a *apiData) getTrack() (utils.TrackInfo, error) {
	params := url.Values{"url": {a.Query}}
	resp, err := a.apiRequest("/api/track", params)
	if err != nil {
		slog.Warn("GetTrack request failed", "error", err)
		return utils.TrackInfo{}, fmt.Errorf("the GetTrack request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return utils.TrackInfo{}, fmt.Errorf("unexpected status code while fetching the track: %s", resp.Status)
	}

	var data utils.TrackInfo
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		slog.Warn("Failed to decode the GetTrack response", "error", err)
		return utils.TrackInfo{}, fmt.Errorf("failed to decode the GetTrack response: %w", err)
	}

	return data, nil
}

func (a *apiData) apiRequest(path string, params url.Values) (*http.Response, error) {
	if !a.isValid() {
		return nil, errors.New("invalid API configuration")
	}

	urlStr := a.ApiUrl + path
	if a.ApiUrl == "" {
		urlStr = path
	}

	if params == nil {
		params = url.Values{}
	}

	// Prefer header-based API key auth, but fall back to query-based Shruti style.
	headers := map[string]string{"X-API-Key": a.APIKey}
	resp, err := sendRequest(http.MethodGet, fmt.Sprintf("%s?%s", urlStr, params.Encode()), nil, headers)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code") || strings.Contains(err.Error(), "404") {
			fallbackParams := url.Values{}
			for k, v := range params {
				fallbackParams[k] = v
			}
			fallbackParams.Set("api_key", a.APIKey)
			return sendRequest(http.MethodGet, fmt.Sprintf("%s?%s", a.ApiUrl+path, fallbackParams.Encode()), nil, nil)
		}
		return nil, err
	}

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		fallbackParams := url.Values{}
		for k, v := range params {
			fallbackParams[k] = v
		}
		fallbackParams.Set("api_key", a.APIKey)
		return sendRequest(http.MethodGet, fmt.Sprintf("%s?%s", a.ApiUrl+path, fallbackParams.Encode()), nil, nil)
	}

	return resp, nil
}

func (a *apiData) buildDownloadURL(video bool) string {
	downloadType := "audio"
	if video {
		downloadType = "video"
	}

	params := url.Values{
		"type":    {downloadType},
		"url":     {a.Query},
		"api_key": {a.APIKey},
	}

	return fmt.Sprintf("%s/download?%s", strings.TrimRight(a.ApiUrl, "/"), params.Encode())
}

// downloadTrack downloads a track using the API. If the track is a YouTube video and video format is requested,
func (a *apiData) downloadTrack(info utils.TrackInfo, video bool) (string, error) {
	// if the track is from YouTube and video:true
	yt := newYouTubeData(a.Query)
	if info.Platform == utils.YouTube && video {
		return yt.downloadTrack(info, video)
	}

	downloader, err := newDownload(info)
	if err != nil {
		return "", fmt.Errorf("failed to initialize the download: %w", err)
	}

	filePath, err := downloader.Process()
	if err != nil {
		if info.Platform == utils.YouTube {
			return yt.downloadTrack(info, video)
		}
		return "", fmt.Errorf("the download process failed: %w", err)
	}

	if strings.Contains(a.ApiUrl, filePath) {
		return downloadFile(filePath, "", false)
	}

	return filePath, nil
}
