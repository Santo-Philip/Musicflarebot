package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/src/core/cache"
	"musicflarebot/src/core/db"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func authListHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	if IsPrivate(m) {
		return nil
	}

	chatID := m.ChatID()

	authUser := db.Instance.GetAuthUsers(chatID)
	if authUser == nil || len(authUser) == 0 {
		_, _ = m.Reply("No authorized users found.")
		return nil
	}

	text := "<b>Authorized Users</b>\n\n"
	for _, uid := range authUser {
		text += fmt.Sprintf("• <a href=\"tg://user?id=%d\">%d</a>\n", uid, uid)
	}

	_, _ = m.Reply(text, replyOpts)
	return nil
}

func addAuthHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	if IsPrivate(m) {
		return nil
	}

	chatID := m.ChatID()

	botStatus, err := cache.GetUserAdmin(client, chatID, m.SenderID(), false)
	if err != nil {
		slog.Warn("GetUserAdmin error", "error", err)
		_, _ = m.Reply("Unable to verify administrator status.")
		return nil
	}

	switch botStatus.Status {
	case "creator", "administrator":
	default:
		_, _ = m.Reply("You must be an administrator to use this command.")
		return nil
	}

	userID, err := getTargetUserID(m)
	if err != nil {
		_, _ = m.Reply(err.Error())
		return nil
	}

	if db.Instance.IsAuthUser(chatID, userID) {
		_, _ = m.Reply("This user is already authorized.")
		return nil
	}

	if err = db.Instance.AddAuthUser(chatID, userID); err != nil {
		slog.Error("Failed to add authorized user", "error", err)
		_, _ = m.Reply("Failed to authorize the user.")
		return nil
	}

	_, err = m.Reply(fmt.Sprintf("User %d has been authorized.", userID))
	return err
}

func removeAuthHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	if IsPrivate(m) {
		return nil
	}

	chatID := m.ChatID()

	botStatus, err := cache.GetUserAdmin(client, chatID, m.SenderID(), false)
	if err != nil {
		slog.Warn("GetUserAdmin error", "error", err)
		_, _ = m.Reply("Unable to verify administrator status.")
		return nil
	}

	switch botStatus.Status {
	case "creator", "administrator":
	default:
		_, _ = m.Reply("You must be an administrator to use this command.")
		return nil
	}

	userID, err := getTargetUserID(m)
	if err != nil {
		_, _ = m.Reply(err.Error())
		return nil
	}

	if !db.Instance.IsAuthUser(chatID, userID) {
		_, _ = m.Reply("This user is not authorized.")
		return nil
	}

	if err := db.Instance.RemoveAuthUser(chatID, userID); err != nil {
		slog.Error("Failed to remove authorized user", "error", err)
		_, _ = m.Reply("Failed to remove authorized user.")
		return nil
	}

	_, err = m.Reply(fmt.Sprintf("User %d has been removed from the authorized list.", userID))
	return err
}
