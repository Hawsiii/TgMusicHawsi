/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package vc

import (
	"log/slog"
	"regexp"
	"sync"
	"time"

	"ashokshau/tgmusic/src/core/cache"

	td "github.com/AshokShau/gotdbot"
	tg "github.com/amarnathcjd/gogram/telegram"
)

var logger = slog.Default()
var urlRegex = regexp.MustCompile(`^https?://`)

// TelegramCalls manages the state and operations for voice calls, including userbots and the main bot client.
type TelegramCalls struct {
	mu          sync.RWMutex
	assistants  map[int]*Assistant
	clients     map[int]*tg.Client
	statusCache *cache.Cache[td.ChatMemberStatus]
	inviteCache *cache.Cache[string]

	replacing map[int64]bool
}

var (
	instance *TelegramCalls
	once     sync.Once
)

// getCalls returns the singleton instance of the TelegramCalls manager, ensuring that only one instance is created.
func getCalls() *TelegramCalls {
	once.Do(func() {
		instance = &TelegramCalls{
			assistants:  make(map[int]*Assistant),
			clients:     make(map[int]*tg.Client),
			statusCache: cache.NewCache[td.ChatMemberStatus](2 * time.Hour),
			inviteCache: cache.NewCache[string](2 * time.Hour),

			replacing: make(map[int64]bool),
		}
	})
	return instance
}

func (c *TelegramCalls) setReplacing(chatID int64, v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v {
		c.replacing[chatID] = true
	} else {
		delete(c.replacing, chatID)
	}
}

func (c *TelegramCalls) isReplacing(chatID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.replacing[chatID]
}

// Calls is the singleton instance of TelegramCalls, initialized lazily.
var Calls = getCalls()
