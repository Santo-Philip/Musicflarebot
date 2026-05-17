package handlers

import (
	"fmt"
	"musicflarebot/src/core"
	"musicflarebot/src/core/cache"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func handleVoiceChatMessage(m *tg.NewMessage) error {
	chatID := m.ChatID()

	if m.ChatID() > 0 {
		text := fmt.Sprintf(
			"This chat (%d) is not a supergroup yet.\n<b>⚠️ Please convert this chat to a supergroup and add me as admin.</b>\n\nIf you don't know how to convert, use this guide:\n🔗 https://te.legra.ph/How-to-Convert-a-Group-to-a-Supergroup-01-02\n\nIf you have any questions, join our support group:",
			chatID,
		)

		_, _ = client.SendMessage(chatID, text, &tg.SendOptions{
			ReplyMarkup: core.AddMeMarkup(client.Me().Username),
			LinkPreview: false,
			ParseMode:   "HTML",
		})

		time.Sleep(1 * time.Second)
		_ = client.LeaveChannel(chatID)
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
