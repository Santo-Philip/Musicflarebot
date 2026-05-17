package handlers

import (
	"log/slog"
	"musicflarebot/src/utils"
	"slices"
	"strings"

	"musicflarebot/src/core/cache"
	"musicflarebot/src/core/db"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func checkBotAdmin(m *tg.NewMessage) bool {
	chatID := m.ChatID()
	botStatus, err := cache.GetUserAdmin(client, chatID, client.Me().ID, false)
	if err != nil {
		if strings.Contains(err.Error(), "is not an admin in chat") {
			_, _ = m.Reply("Bot is not an administrator in this chat. Please promote the bot with invite users permission.")
		} else {
			slog.Warn("GetUserAdmin error", "error", err)
			_, _ = m.Reply("Unable to verify bot administrator status.")
		}
		return false
	}

	switch botStatus.Status {
	case "creator":
		return true
	case "administrator":
		if botStatus.Rights == nil || !botStatus.Rights.InviteUsers {
			_, _ = m.Reply("The bot does not have permission to invite users.")
			return false
		}
		return true
	default:
		_, _ = m.Reply("Bot is not an administrator in this chat. Use /reload to refresh admin cache.")
		return false
	}
}

func adminMode(m *tg.NewMessage) bool {
	if IsPrivate(m) {
		return false
	}

	chatID := m.ChatID()

	if !checkBotAdmin(m) {
		return false
	}

	userID := m.SenderID()
	switch db.Instance.GetAdminMode(chatID) {
	case utils.Everyone:
		return true
	case utils.Admins:
		if db.Instance.IsAdmin(chatID, userID) || db.Instance.IsAuthUser(chatID, userID) {
			return true
		}
		_, _ = m.Reply("You must be an administrator to use this command.")
		return false
	default:
		_, _ = m.Reply("You are not authorized to use this command.")
		return false
	}
}

func adminModeCB(q *tg.CallbackQuery) bool {
	chatID := q.ChatID

	if !checkBotAdminCB(q) {
		return false
	}

	userID := q.SenderID
	switch db.Instance.GetAdminMode(chatID) {
	case utils.Everyone:
		return true
	case utils.Admins:
		if db.Instance.IsAdmin(chatID, userID) || db.Instance.IsAuthUser(chatID, userID) {
			return true
		}
		_, _ = q.Answer("You must be an administrator to use this action.", &tg.CallbackOptions{Alert: true})
		return false
	default:
		_, _ = q.Answer("You are not authorized to use this action.", &tg.CallbackOptions{Alert: true})
		return false
	}
}

func checkBotAdminCB(q *tg.CallbackQuery) bool {
	chatID := q.ChatID
	botStatus, err := cache.GetUserAdmin(client, chatID, client.Me().ID, false)
	if err != nil {
		if strings.Contains(err.Error(), "is not an admin in chat") {
			_, _ = q.Answer("Bot is not an administrator in this chat.", &tg.CallbackOptions{Alert: true})
		} else {
			_, _ = q.Answer("Unable to verify bot administrator status.", &tg.CallbackOptions{Alert: true})
		}
		return false
	}

	switch botStatus.Status {
	case "creator", "administrator":
		return true
	default:
		_, _ = q.Answer("Bot is not an administrator in this chat.", &tg.CallbackOptions{Alert: true})
		return false
	}
}

func playMode(m *tg.NewMessage) bool {
	if IsPrivate(m) {
		return false
	}

	chatID := m.ChatID()

	if !checkBotAdmin(m) {
		return false
	}

	if db.Instance.GetPlayMode(chatID) {
		admins, err := cache.GetAdmins(client, chatID, false)
		if err != nil {
			return false
		}

		senderID := m.SenderID()
		isAdmin := slices.ContainsFunc(admins, func(a *tg.Participant) bool {
			return a.User != nil && a.User.ID == senderID
		})

		if !isAdmin && !db.Instance.IsAuthUser(chatID, senderID) {
			_, _ = m.Reply("Play mode is enabled. Only administrators and authorized users can start playback.")
			return false
		}
	}

	return true
}
