package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/src/utils"
	"strings"

	"musicflarebot/src/core"
	"musicflarebot/src/core/cache"
	"musicflarebot/src/core/db"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func settingsHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()
	admins, err := cache.GetAdmins(client, chatID, false)
	if err != nil {
		return err
	}

	var isAdmin bool
	for _, admin := range admins {
		if admin.User != nil && admin.User.ID == m.SenderID() {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return nil
	}

	getPlayMode := db.Instance.GetPlayMode(chatID)
	playModeStr := utils.Everyone
	if getPlayMode {
		playModeStr = utils.Admins
	}
	getAdminMode := db.Instance.GetAdminMode(chatID)
	cmdDelete := db.Instance.GetCmdDelete(chatID)
	language, _ := db.Instance.GetLanguage(chatID)

	chat, err := client.GetChat(chatID)
	if err != nil {
		slog.Warn("Failed to get chat", "error", err)
		return nil
	}

	chatTitle := chat.Title
	if chatTitle == "" {
		chatTitle = fmt.Sprintf("Chat %d", chatID)
	}

	text := fmt.Sprintf("<u><b>%s settings</b></u>\n\nClick the buttons below to change this chat's current settings.",
		chatTitle)

	_, err = m.Reply(text, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.SettingsKeyboard(playModeStr, getAdminMode, cmdDelete, language),
	})
	return err
}

func settingsCallbackHandler(q *tg.CallbackQuery) error {
	chatID := q.ChatID

	admins, err := cache.GetAdmins(client, chatID, false)
	if err != nil {
		return err
	}

	var hasPerms bool
	for _, admin := range admins {
		if admin.User != nil && admin.User.ID == q.SenderID {
			rights, _ := cache.GetRights(client, chatID, q.SenderID, false)
			var isCreator bool = admin.Status == "creator"
			hasPerms = (rights != nil && rights.ManageCall) || isCreator
			break
		}
	}

	if !hasPerms {
		_, _ = q.Answer("You don't have permission to change settings.", &tg.CallbackOptions{Alert: true})
		return nil
	}

	data := q.DataString()
	if data == "settings_main" {
		_, _ = q.Answer("Update your chat settings")
		return nil
	}

	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		return nil
	}

	settingType := parts[1]

	switch settingType {
	case "delete":
		cmdDelete := db.Instance.GetCmdDelete(chatID)
		_ = db.Instance.SetCmdDelete(chatID, !cmdDelete)
	case "play":
		getPlayMode := db.Instance.GetPlayMode(chatID)
		_ = db.Instance.SetPlayMode(chatID, !getPlayMode)
	case "admin":
		getAdminMode := db.Instance.GetAdminMode(chatID)
		newMode := utils.Everyone
		if getAdminMode == utils.Everyone {
			newMode = utils.Admins
		}
		_ = db.Instance.SetAdminMode(chatID, newMode)
	case "lang":
		_, _ = q.Answer("Language selection is not yet implemented via this menu.", &tg.CallbackOptions{Alert: true})
		return nil
	default:
		_, _ = q.Answer("Unknown setting", &tg.CallbackOptions{Alert: true})
		return nil
	}

	getPlayMode := db.Instance.GetPlayMode(chatID)
	playModeStr := utils.Everyone
	if getPlayMode {
		playModeStr = utils.Admins
	}
	getAdminMode := db.Instance.GetAdminMode(chatID)
	cmdDelete := db.Instance.GetCmdDelete(chatID)
	language, _ := db.Instance.GetLanguage(chatID)

	chat, err := client.GetChat(chatID)
	if err != nil {
		slog.Warn("Failed to get chat", "error", err)
		return nil
	}

	chatTitle := chat.Title
	if chatTitle == "" {
		chatTitle = fmt.Sprintf("Chat %d", chatID)
	}

	text := fmt.Sprintf("<u><b>%s settings</b></u>\n\nClick the buttons below to change this chat's current settings.",
		chatTitle)

	_, err = q.Edit(text, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.SettingsKeyboard(playModeStr, getAdminMode, cmdDelete, language),
	})
	if err != nil {
		return err
	}

	_, _ = q.Answer("Settings updated")
	return nil
}
