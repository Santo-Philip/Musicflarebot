package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/src/utils"
	"musicflarebot/src/vc"
	"time"

	"musicflarebot/src/core/cache"

	tg "github.com/amarnathcjd/gogram/telegram"
)

const reloadCooldown = 3 * time.Minute

var reloadRateLimit = cache.NewCache[time.Time](reloadCooldown)

func reloadAdminCacheHandler(m *tg.NewMessage) error {
	if IsPrivate(m) {
		return nil
	}

	reloadKey := fmt.Sprintf("reload:%d", m.ChatID())
	if lastUsed, ok := reloadRateLimit.Get(reloadKey); ok {
		timePassed := time.Since(lastUsed)
		if timePassed < reloadCooldown {
			remaining := int((reloadCooldown - timePassed).Seconds())
			_, _ = m.Reply(fmt.Sprintf("Please wait %s before using this command again.", utils.SecToMin(remaining)))
			return nil
		}
	}

	reloadRateLimit.Set(reloadKey, time.Now())

	reply, err := m.Reply("Reloading administrator cache...")
	if err != nil {
		slog.Warn("Failed to send reloading message for chat", "chat_id", m.ChatID(), "error", err)
		return nil
	}

	cache.ClearAdminCache(m.ChatID())
	vc.Calls.UpdateInviteLink(m.ChatID(), "")

	admins, err := cache.GetAdmins(client, m.ChatID(), true)
	if err != nil {
		slog.Warn("Failed to reload the admin cache for chat", "chat_id", m.ChatID(), "error", err)
		_, _ = reply.Edit("Failed to reload administrator cache.")
		return nil
	}

	slog.Info("Reloaded admins for chat", "count", len(admins), "chat_id", m.ChatID())
	_, _ = reply.Edit("Administrator cache reloaded successfully.")
	return nil
}

func privacyHandler(m *tg.NewMessage) error {
	botName := client.Me().FirstName

	text := fmt.Sprintf("<b>Privacy Policy for %s</b>\n\n<b>1. Data Storage:</b>\nWe do not store personal data on your device. We do not track your browsing activity.\n\n<b>2. Collection:</b>\nWe only collect your Telegram <b>User ID</b> and <b>Chat ID</b> to provide music services. No names, phone numbers, or locations are stored.\n\n<b>3. Usage:</b>\nData is used strictly for bot functionality. No marketing or commercial use.\n\n<b>4. Sharing:</b>\nWe do not share data with third parties. No data is sold or traded.\n\n<b>5. Security:</b>\nWe use standard encryption to protect data. However, no online service is 100%% secure.\n\n<b>6. Cookies:</b>\n%s does not use cookies or tracking technologies.\n\n<b>7. Third Parties:</b>\nWe do not integrate with third-party data collectors, other than Telegram itself.\n\n<b>8. Your Rights:</b>\nYou can request data deletion or block the bot to revoke access.\n\n<b>9. Updates:</b>\nPolicy changes will be announced in the bot.\n\n<b>10. Contact:</b>\nQuestions? Contact our <a href=\"https://t.me/GuardxSupport\">Support Group</a>.\n\n──────────────────\n<b>Note:</b> This policy ensures a safe and respectful experience with %s.", botName, botName, botName)

	_, err := m.Reply(text, &tg.SendOptions{ParseMode: "html", LinkPreview: false})
	return err
}
