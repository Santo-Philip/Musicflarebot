package handlers

import (
	"musicflarebot/internal/cache"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func skipHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, _ = m.Reply("The bot is not streaming in the video chat.")
		return nil
	}

	_ = vc.Calls.PlayNext(chatID)
	return nil
}
