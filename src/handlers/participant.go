package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/internal/config"
	"musicflarebot/internal/cache"
	"musicflarebot/internal/database"
	"musicflarebot/src/vc"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func handleParticipant(pu *tg.ParticipantUpdate) error {
	chatID := pu.ChatID()
	userID := pu.UserID()
	me := client.Me()

	call, _, err := vc.Calls.GetGroupAssistant(chatID)
	if err != nil {
		return nil
	}

	ubID := call.App.Me().ID

	if !isRelevantUser(userID, me.ID, ubID) {
		return nil
	}

	s := strings.TrimPrefix(strconv.FormatInt(chatID, 10), "-100")
	rawChatID, _ := strconv.ParseInt(s, 10, 64)

	gChat, err := client.GetChannel(rawChatID)
	if err != nil {
		if strings.Contains(err.Error(), "CHANNEL_INVALID") || strings.Contains(err.Error(), "Invalid supergroup identifier") {
			_ = client.LeaveChannel(chatID)
			return nil
		}
		slog.Error("Failed to get channel", "chat_id", chatID, "error", err)
		return nil
	}

	if gChat.Username != "" {
		inviteLink := fmt.Sprintf("https://t.me/%s", gChat.Username)
		vc.Calls.UpdateInviteLink(chatID, inviteLink)
	}

	go storeChatReference(chatID)

	oldStatus := participantStatusToString(pu.Old)
	newStatus := participantStatusToString(pu.New)

	if isAdminStatus(oldStatus) || isAdminStatus(newStatus) {
		newParticipant := &tg.Participant{
			User: &tg.UserObj{ID: userID},
		}
		if pu.New != nil {
			newParticipant.Status = newStatus
			newParticipant.Rights = participantRights(pu.New)
		}
		cache.UpdateAdminCache(chatID, newParticipant)
	}

	slog.Debug("Status change", "user_id", userID, "old", oldStatus, "new", newStatus, "chat_id", chatID)

	return handleParticipantStatusChange(chatID, userID, ubID, oldStatus, newStatus)
}

func participantStatusToString(p tg.ChannelParticipant) string {
	if p == nil {
		return "left"
	}
	switch p.(type) {
	case *tg.ChannelParticipantCreator:
		return "creator"
	case *tg.ChannelParticipantAdmin:
		return "administrator"
	case *tg.ChannelParticipantBanned:
		return "kicked"
	case *tg.ChannelParticipantLeft:
		return "left"
	default:
		return "member"
	}
}

func participantRights(p tg.ChannelParticipant) *tg.ChatAdminRights {
	switch pt := p.(type) {
	case *tg.ChannelParticipantAdmin:
		return pt.AdminRights
	default:
		return nil
	}
}

func isAdminStatus(status string) bool {
	return status == "administrator" || status == "creator"
}

func storeChatReference(chatID int64) {
	slog.Debug("Storing chat reference for chat", "chat_id", chatID)

	if err := database.AddChat(chatID); err != nil {
		slog.Error("Failed to add chat to database", "chat_id", chatID, "error", err)
	}
}

func isRelevantUser(userID, botID, assistantID int64) bool {
	return userID == botID || userID == assistantID
}

func handleParticipantStatusChange(chatID int64, userID int64, ubID int64, oldStatus string, newStatus string) error {
	switch {
	case oldStatus == "left" && (newStatus == "member" || newStatus == "administrator" || newStatus == "creator"):
		return handleJoinG(chatID, userID, ubID)

	case (oldStatus == "member" || oldStatus == "administrator") && newStatus == "left":
		return handleLeaveG(chatID, userID, ubID)

	case newStatus == "kicked":
		return handleBanG(chatID, userID, ubID)

	case oldStatus == "kicked" && newStatus == "left":
		return handleUnbanG(chatID, userID)

	default:
		return handlePromotionDemotionG(chatID, userID, oldStatus, newStatus)
	}
}

func handleJoinG(chatID int64, userID int64, ubID int64) error {
	slog.Info("User joined chat", "user_id", userID, "chat_id", chatID)

	if userID == client.Me().ID {
		slog.Info("Bot joined chat", "chat_id", chatID)
		sendJoinLogG(chatID)
	}

	vc.Calls.UpdateMembership(chatID, userID, "member")
	return nil
}

func sendJoinLogG(chatID int64) {
	text := fmt.Sprintf(
		"<b>🤖 Bot Joined a New Chat</b>\n"+
			"📌 <b>Chat ID:</b> <code>%d</code>\n",
		chatID,
	)

	_, err := client.SendMessage(
		config.Conf.LoggerId,
		text,
		&tg.SendOptions{ParseMode: "HTML"},
	)

	if err != nil {
		slog.Warn("Failed to send join log", "error", err)
	}
}

func handleLeaveG(chatID int64, userID int64, ubID int64) error {
	slog.Info("User left chat", "user_id", userID, "chat_id", chatID)

	if userID == ubID {
		cache.ChatCache.ClearChat(chatID)
	}

	if userID == client.Me().ID {
		if err := vc.Calls.Stop(chatID); err != nil {
			slog.Error("Failed to stop VC", "error", err)
		}
	}

	vc.Calls.UpdateMembership(chatID, userID, "left")
	return nil
}

func handleBanG(chatID int64, userID int64, ubID int64) error {
	slog.Debug("User banned from chat", "user_id", userID, "chat_id", chatID)

	if userID == ubID {
		cache.ChatCache.ClearChat(chatID)

		message := fmt.Sprintf(
			"🚫 My assistant has been banned from this chat.\n\n"+
				"If this was a mistake please unban <code>%d</code>.",
			ubID,
		)

		_, err := client.SendMessage(
			chatID,
			message,
			&tg.SendOptions{ParseMode: "HTML"},
		)

		if err != nil {
			return err
		}
	}

	if userID == client.Me().ID {
		if err := vc.Calls.Stop(chatID); err != nil {
			slog.Error("Failed stopping VC after ban", "error", err)
		}
	}

	vc.Calls.UpdateMembership(chatID, userID, "kicked")
	return nil
}

func handleUnbanG(chatID int64, userID int64) error {
	slog.Info("User unbanned from chat", "user_id", userID, "chat_id", chatID)
	vc.Calls.UpdateMembership(chatID, userID, "left")
	return nil
}

func handlePromotionDemotionG(chatID int64, userID int64, oldStatus string, newStatus string) error {
	oldAdmin := isAdminStatus(oldStatus)
	newAdmin := isAdminStatus(newStatus)

	isPromoted := !oldAdmin && newAdmin
	isDemoted := oldAdmin && !newAdmin

	if !isPromoted && !isDemoted {
		return nil
	}

	if isPromoted {
		slog.Info("User promoted in chat", "user_id", userID, "chat_id", chatID)
		vc.Calls.UpdateMembership(chatID, userID, "administrator")
		return nil
	}

	slog.Info("User demoted in chat", "user_id", userID, "chat_id", chatID)
	vc.Calls.UpdateMembership(chatID, userID, "member")
	return nil
}
