package handlers

import (
	"fmt"

	"musicflarebot/src/core/cache"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func stopHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, _ = m.Reply("The bot isn't streaming in the video chat.")
		return nil
	}

	_ = vc.Calls.Stop(chatID)
	_, _ = m.Reply(fmt.Sprintf("<b>Stream ended by</b> %s", firstName(m)), replyOpts)
	return nil
}
