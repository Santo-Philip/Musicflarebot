package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/internal/database"
	"strconv"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func myStatHandler(m *tg.NewMessage) error {
	userID := m.SenderID()
	chatID := m.ChatID()

	totalPlays, err := database.GetUserTotalPlays(userID)
	if err != nil {
		slog.Warn("[myStat] GetUserTotalPlays error", "error", err)
		totalPlays = 0
	}

	globalRank, _ := database.GetUserGlobalRank(userID)

	userLink := fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", userID, firstName(m))

	text := fmt.Sprintf(
		"<b>📊 Your Music Stats</b>\n\n"+
			"<b>User:</b> %s\n"+
			"<b>Global Plays:</b> %d\n"+
			"<b>Global Rank:</b> #%d\n",
		userLink, totalPlays, globalRank,
	)

	if IsGroup(m) {
		groupPlays, err := database.GetUserGroupPlays(userID, chatID)
		if err != nil {
			groupPlays = 0
		}
		text += fmt.Sprintf("<b>Group Plays:</b> %d\n", groupPlays)
	}

	_, err = m.Reply(text, &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
	return err
}

func groupStatHandler(m *tg.NewMessage) error {
	if IsPrivate(m) {
		_, err := m.Reply("This command only works in groups.")
		return err
	}

	chatID := m.ChatID()

	top, err := database.GetGroupTop(chatID, 25)
	if err != nil {
		_, err = m.Reply("Failed to fetch group stats.")
		return err
	}

	if len(top) == 0 {
		_, err = m.Reply("No stats available for this group yet.")
		return err
	}

	chat, err := client.GetChat(chatID)
	if err != nil {
		slog.Warn("[groupStat] GetChat error", "error", err)
	}

	chatName := "this group"
	if chat != nil {
		chatName = chat.Title
	}

	text := fmt.Sprintf("<b>🏆 Top 25 in %s</b>\n\n", chatName)

	for i, stat := range top {
		rank := i + 1
		medal := ""
		switch rank {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		default:
			medal = strconv.Itoa(rank)
		}

		u, err := client.GetUser(stat.UserID)
		userDisplay := fmt.Sprintf("User #%d", stat.UserID)
		if err == nil && u != nil {
			name := u.FirstName
			if u.LastName != "" {
				name += " " + u.LastName
			}
			userDisplay = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", stat.UserID, name)
		}

		text += fmt.Sprintf("<b>%s.</b> %s — <b>%d</b> plays\n", medal, userDisplay, stat.Plays)
	}

	_, err = m.Reply(text, &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
	return err
}

func globalStatHandler(m *tg.NewMessage) error {
	userID := m.SenderID()

	top, err := database.GetGlobalTop(25)
	if err != nil {
		_, err = m.Reply("Failed to fetch global stats.")
		return err
	}

	if len(top) == 0 {
		_, err = m.Reply("No stats available yet.")
		return err
	}

	text := "<b>🌍 Global Top 25</b>\n\n"

	for i, stat := range top {
		rank := i + 1
		medal := ""
		switch rank {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		default:
			medal = strconv.Itoa(rank)
		}

		u, err := client.GetUser(stat.UserID)
		userDisplay := fmt.Sprintf("User #%d", stat.UserID)
		if err == nil && u != nil {
			name := u.FirstName
			if u.LastName != "" {
				name += " " + u.LastName
			}
			userDisplay = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", stat.UserID, name)
		}

		text += fmt.Sprintf("<b>%s.</b> %s — <b>%d</b> plays\n", medal, userDisplay, stat.Plays)
	}

	globalRank, err := database.GetUserGlobalRank(userID)
	if err == nil && globalRank > 0 {
		totalPlays, _ := database.GetUserTotalPlays(userID)
		text += fmt.Sprintf("\n<b>Your Ranking:</b> #%d with %d plays", globalRank, totalPlays)
	}

	_, err = m.Reply(text, &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
	return err
}
