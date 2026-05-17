package handlers

import (
	"musicflarebot/src/core/cache"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func handleVoiceChatMessage(m *tg.NewMessage) error {
	chatID := m.ChatID()

	if IsPrivate(m) || chatID > 0 {
		return nil
	}

	if m.Action != nil {
		switch m.Action.(type) {
		case *tg.MessageActionGroupCall:
			cache.ChatCache.ClearChat(chatID)
			_, _ = client.SendMessage(chatID, "🎙️ Video chat started!\nUse /play <song name> to play music.")
		case *tg.MessageActionGroupCallScheduled:
		}
	}

	return nil
}
