package handlers

import (
	"log/slog"
	"musicflarebot/internal/types"
	"slices"
	"strings"

	"musicflarebot/internal/cache"
	"musicflarebot/internal/database"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func checkBotAdmin(m *tg.NewMessage) bool {
	chatID := m.ChatID()
	botID := client.Me().ID
	slog.Debug("checkBotAdmin", "chatID", chatID, "botID", botID)
	botStatus, err := cache.GetUserAdmin(client, chatID, botID, false)
	if err != nil {
		slog.Warn("checkBotAdmin: GetUserAdmin error", "error", err, "chatID", chatID, "botID", botID)
		if strings.Contains(err.Error(), "is not an admin in chat") {
			_, _ = m.Reply("Bot is not an administrator in this chat. Please promote the bot with invite users permission.")
		} else {
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

	if IsSuperGroup(m) {
		if !checkBotAdmin(m) {
			return false
		}
	}

	userID := m.SenderID()
	switch database.GetAdminMode(chatID) {
	case types.Everyone:
		return true
	case types.Admins:
		if database.IsAdmin(chatID, userID) || database.IsAuthUser(chatID, userID) {
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

	if chatID > 0 {
		return false
	}

	if (chatID < -1000000000 || chatID > -1000000000) && !checkBotAdminCB(q) {
		return false
	}

	userID := q.SenderID
	switch database.GetAdminMode(chatID) {
	case types.Everyone:
		return true
	case types.Admins:
		if database.IsAdmin(chatID, userID) || database.IsAuthUser(chatID, userID) {
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
	chatID := m.ChatID()

	// Resolve the peer type to reliably detect private chat vs any group.
	// Basic groups can return positive ChatID() (same as private chats), so we
	// cannot rely on numeric sign alone. Using GetInputPeer avoids that ambiguity.
	peer, err := client.GetInputPeer(chatID)
	if err == nil {
		if _, ok := peer.(*tg.InputPeerUser); ok {
			_, _ = m.Reply("Playback commands work only in group voice chats.")
			return false
		}
	} else {
		// Fallback: if GetInputPeer fails, check via ChatID vs SenderID.
		if m.ChatID() > 0 && m.SenderID() == m.ChatID() {
			_, _ = m.Reply("Playback commands work only in group voice chats.")
			return false
		}
	}

	if database.GetPlayMode(chatID) {
		admins, err := cache.GetAdmins(client, chatID, false)
		if err != nil {
			return false
		}

		senderID := m.SenderID()
		isAdmin := slices.ContainsFunc(admins, func(a *tg.Participant) bool {
			return a.User != nil && a.User.ID == senderID
		})

		if !isAdmin && !database.IsAuthUser(chatID, senderID) {
			_, _ = m.Reply("Play mode is enabled. Only administrators and authorized users can start playback.")
			return false
		}
	}

	return true
}
