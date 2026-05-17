package handlers

import (
	"errors"
	"fmt"
	"musicflarebot/config"
	"strconv"
	"strings"
	"time"

	"log/slog"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func Args(m *tg.NewMessage) string {
	parts := strings.Split(m.Text(), " ")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(strings.Join(parts[1:], " "))
}

func firstName(m *tg.NewMessage) string {
	user, err := client.GetUser(m.SenderID())
	if err != nil {
		return "Unknown"
	}
	return user.FirstName
}

var replyOpts = &tg.SendOptions{
	ParseMode:   "HTML",
	LinkPreview: false,
}

func isDev(m *tg.NewMessage) bool {
	userID := m.SenderID()
	for _, dev := range config.Conf.DEVS {
		if dev == userID {
			return true
		}
	}
	return false
}

func IsPrivate(m *tg.NewMessage) bool {
	return m.ChatID() > 0
}

func getTargetUserID(m *tg.NewMessage) (int64, error) {
	if m.ReplyToMsgID() != 0 {
		return resolveFromReply(m)
	}

	args := strings.Fields(Args(m))
	if len(args) == 0 {
		return 0, errors.New("no target specified: reply to a message or provide a user ID/username")
	}

	userID, err := resolveFromArg(args[0])
	if err != nil {
		return 0, err
	}

	if m.SenderID() == userID {
		return 0, errors.New("cannot perform action on yourself")
	}

	return userID, nil
}

func getReplyMessage(m *tg.NewMessage) (*tg.NewMessage, error) {
	if m.ReplyToMsgID() == 0 {
		return nil, errors.New("no reply")
	}
	return client.GetMessageByID(m.ChatID(), m.ReplyToMsgID())
}

func resolveFromReply(m *tg.NewMessage) (int64, error) {
	replyMsg, err := getReplyMessage(m)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch replied message: %w", err)
	}

	userID := replyMsg.SenderID()
	if userID == 0 {
		return 0, errors.New("replied message has no identifiable sender")
	}

	return userID, nil
}

func resolveFromArg(arg string) (int64, error) {
	if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
		if id <= 0 {
			return 0, fmt.Errorf("invalid user ID: %d", id)
		}
		return id, nil
	}

	return resolveUsername(arg)
}

func resolveUsername(username string) (int64, error) {
	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return 0, errors.New("username cannot be empty")
	}

	result, err := client.ResolveUsername(username)
	if err != nil {
		slog.Warn("username lookup failed", "username", username, "error", err)
		return 0, fmt.Errorf("username lookup failed for %q: %w", username, err)
	}

	if result == nil {
		return 0, fmt.Errorf("no user found for username %q", username)
	}

	switch peer := result.(type) {
	case *tg.InputPeerUser:
		return peer.UserID, nil
	case *tg.InputPeerChannel:
		return peer.ChannelID, nil
	case *tg.InputPeerChat:
		return peer.ChatID, nil
	default:
		return 0, fmt.Errorf("unknown peer type for username %q", username)
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func getFormattedDuration(diff time.Duration) string {
	totalSeconds := int(diff.Seconds())

	months := totalSeconds / (30 * 24 * 3600)
	remaining := totalSeconds % (30 * 24 * 3600)

	weeks := remaining / (7 * 24 * 3600)
	remaining = remaining % (7 * 24 * 3600)

	days := remaining / (24 * 3600)
	remaining = remaining % (24 * 3600)

	hours := remaining / 3600
	remaining = remaining % 3600

	minutes := remaining / 60
	seconds := remaining % 60

	var parts []string

	if months > 0 {
		parts = append(parts, plural(months, "month"))
	}
	if weeks > 0 {
		parts = append(parts, plural(weeks, "week"))
	}
	if days > 0 {
		parts = append(parts, plural(days, "day"))
	}
	if hours > 0 {
		parts = append(parts, plural(hours, "hour"))
	}
	if minutes > 0 {
		parts = append(parts, plural(minutes, "minute"))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, plural(seconds, "second"))
	}

	return strings.Join(parts, " ")
}
