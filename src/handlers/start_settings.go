package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/config"
	"musicflarebot/src/core/db"
	"os"
	"path/filepath"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func setStartMessageHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /setstartmsg <message>\n\nSend the custom start message with HTML formatting.")
		return nil
	}

	if err := db.Instance.SetSetting(db.SettingStartMessage, args); err != nil {
		slog.Error("Failed to set start message", "error", err)
		_, _ = m.Reply("Failed to save start message.")
		return nil
	}

	_, _ = m.Reply("Start message has been updated.\n\n<code>Preview:</code>\n"+args, &tg.SendOptions{ParseMode: "HTML"})
	return nil
}

func setStartMediaHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	loggerID := config.Conf.LoggerId
	if loggerID == 0 {
		_, _ = m.Reply("No log channel set. Configure one with /set_logger_id first.")
		return nil
	}

	if m.ReplyToMsgID() == 0 {
		_, _ = m.Reply("Reply to a media message (photo, video, GIF, or document) to set as start media.")
		return nil
	}

	reply, err := getReplyMessage(m)
	if err != nil || !isValidMedia(reply) {
		_, _ = m.Reply("Reply to a valid media message.")
		return nil
	}

	path, err := reply.Download()
	if err != nil {
		_, _ = m.Reply("Failed to download media.")
		return nil
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		ext = ".dat"
	}
	dest := filepath.Join(config.Conf.DownloadsDir, "start_custom"+ext)
	if err := os.Rename(path, dest); err != nil {
		if err := copyFile(path, dest); err != nil {
			os.Remove(path)
			_, _ = m.Reply("Failed to save media.")
			return nil
		}
		os.Remove(path)
	}

	if err := db.Instance.SetSetting(db.SettingStartMedia, dest); err != nil {
		os.Remove(dest)
		_, _ = m.Reply("Failed to save start media path.")
		return nil
	}

	_, fwdErr := client.Forward(loggerID, m.ChatID(), []int32{m.ReplyToMsgID()})
	if fwdErr != nil {
		slog.Warn("[setStartMedia] Failed to forward to log channel", "error", fwdErr)
	} else {
		slog.Info("[setStartMedia] Media forwarded to log channel", "loggerID", loggerID)
	}

	_, _ = m.Reply(fmt.Sprintf("Start media has been set and saved to log channel [%d].", loggerID))
	return nil
}

func resetStartHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	customPath, _ := db.Instance.GetSetting(db.SettingStartMedia)
	if customPath != "" {
		os.Remove(customPath)
	}

	_ = db.Instance.DeleteSetting(db.SettingStartMessage)
	_ = db.Instance.DeleteSetting(db.SettingStartMedia)
	_ = db.Instance.DeleteSetting(db.SettingStartMediaType)

	_, _ = m.Reply("Start message and media have been reset to defaults.")
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
