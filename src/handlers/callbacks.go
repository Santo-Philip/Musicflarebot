package handlers

import (
	"fmt"
	"html"
	"log/slog"
	"musicflarebot/src/utils"
	"strings"

	"musicflarebot/src/core"
	"musicflarebot/src/core/cache"
	"musicflarebot/src/core/db"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func playCallbackHandler(q *tg.CallbackQuery) error {
	if !adminModeCB(q) {
		return nil
	}

	data := q.DataString()
	if strings.Contains(data, "settings_") {
		return nil
	}

	chatID := q.ChatID
	user, err := client.GetUser(q.SenderID)
	if err != nil {
		user = &tg.UserObj{FirstName: "Unknown", ID: q.SenderID}
	}

	if !cache.ChatCache.IsActive(chatID) {
		text := "There is no active playback."
		_, _ = q.Answer(text)
		_, _ = q.Edit(text, &tg.SendOptions{ReplyMarkup: core.ControlButtons(""), ParseMode: "HTML", LinkPreview: false})
		return nil
	}

	currentTrack := cache.ChatCache.GetPlayingTrack(chatID)
	if currentTrack == nil {
		_, _ = q.Answer("There is no active playback.")
		_, _ = q.Edit("There is no active playback.", &tg.SendOptions{ReplyMarkup: core.ControlButtons(""), ParseMode: "HTML", LinkPreview: false})
		return nil
	}

	buildTrackMessage := func(status, emoji string) string {
		escURL := html.EscapeString(currentTrack.URL)
		escName := html.EscapeString(currentTrack.Name)
		escUser := html.EscapeString(currentTrack.User)
		return fmt.Sprintf("%s <b>%s</b>\n\n<b>Track:</b> <a href='%s'>%s</a>\n<b>Duration:</b> %s\n<b>Requested by:</b> %s",
			emoji, status,
			escURL, escName,
			utils.SecToMin(currentTrack.Duration),
			escUser,
		)
	}

	switch {
	case strings.Contains(data, "play_skip"):
		if err := vc.Calls.PlayNext(chatID); err != nil {
			_, _ = q.Answer("Unable to skip the current track.")
			_, _ = q.Edit("Unable to skip the current track.", &tg.SendOptions{ReplyMarkup: core.ControlButtons(""), ParseMode: "HTML", LinkPreview: false})
			return nil
		}
		_, _ = q.Answer("Track skipped.")
		_, _ = client.DeleteMessages(chatID, []int32{q.MessageID})
		return nil

	case strings.Contains(data, "play_stop"):
		if err := vc.Calls.Stop(chatID); err != nil {
			_, _ = q.Answer("Unable to stop playback.")
			_, _ = q.Edit("Unable to stop playback.", &tg.SendOptions{ReplyMarkup: core.ControlButtons(""), ParseMode: "HTML", LinkPreview: false})
			return nil
		}

		msg := fmt.Sprintf("<b>Playback stopped.</b>\nRequested by: %s", html.EscapeString(user.FirstName))
		_, _ = q.Answer("Playback stopped.")
		_, err := q.Edit(msg, &tg.SendOptions{ReplyMarkup: core.ControlButtons(""), ParseMode: "HTML", LinkPreview: false})
		return err

	case strings.Contains(data, "play_pause"):
		if _, err = vc.Calls.Pause(chatID); err != nil {
			_, _ = q.Answer("Unable to pause playback.")
			_, _ = q.Edit("Unable to pause playback.", &tg.SendOptions{ReplyMarkup: core.ControlButtons(""), ParseMode: "HTML", LinkPreview: false})
			return nil
		}
		_, _ = q.Answer("Playback paused.")
		text := buildTrackMessage("Paused", "⏸") + fmt.Sprintf("\n\nPaused by %s", html.EscapeString(user.FirstName))
		_, _ = q.Edit(text, &tg.SendOptions{ReplyMarkup: core.ControlButtons("pause"), ParseMode: "HTML", LinkPreview: false})
		return nil

	case strings.Contains(data, "play_resume"):
		if _, err := vc.Calls.Resume(chatID); err != nil {
			_, _ = q.Answer("Unable to resume playback.")
			_, _ = q.Edit("Unable to resume playback.", &tg.SendOptions{ReplyMarkup: core.ControlButtons("pause"), ParseMode: "HTML", LinkPreview: false})
			return nil
		}
		_, _ = q.Answer("Playback resumed.")
		text := buildTrackMessage("Now Playing", "▶") + fmt.Sprintf("\n\nResumed by %s", html.EscapeString(user.FirstName))
		_, _ = q.Edit(text, &tg.SendOptions{ReplyMarkup: core.ControlButtons("resume"), ParseMode: "HTML", LinkPreview: false})
		return nil

	case strings.Contains(data, "play_mute"):
		if _, err := vc.Calls.Mute(chatID); err != nil {
			_, _ = q.Answer("Unable to mute playback.")
			_, _ = q.Edit("Unable to mute playback.", &tg.SendOptions{ReplyMarkup: core.ControlButtons("mute"), ParseMode: "HTML", LinkPreview: false})
			return nil
		}
		_, _ = q.Answer("Playback muted.")
		text := buildTrackMessage("Muted", "🔇") + fmt.Sprintf("\n\nMuted by %s", html.EscapeString(user.FirstName))
		_, _ = q.Edit(text, &tg.SendOptions{ReplyMarkup: core.ControlButtons("mute"), ParseMode: "HTML", LinkPreview: false})
		return nil

	case strings.Contains(data, "play_unmute"):
		if _, err := vc.Calls.Unmute(chatID); err != nil {
			_, _ = q.Answer("Unable to unmute playback.")
			_, _ = q.Edit("Unable to unmute playback.", &tg.SendOptions{ReplyMarkup: core.ControlButtons("unmute"), ParseMode: "HTML"})
			return nil
		}
		_, _ = q.Answer("Playback unmuted.")
		text := buildTrackMessage("Now Playing", "▶") + fmt.Sprintf("\n\nUnmuted by %s", html.EscapeString(user.FirstName))
		_, _ = q.Edit(text, &tg.SendOptions{ReplyMarkup: core.ControlButtons("unmute"), LinkPreview: false})
		return nil

	case strings.Contains(data, "play_add_to_list"):
		playlists, err := db.Instance.GetUserPlaylists(q.SenderID)
		if err != nil {
			_, _ = q.Answer("Unable to fetch playlists.")
			return nil
		}

		var playlistID string
		if len(playlists) == 0 {
			playlistID, err = db.Instance.CreatePlaylist("My Playlist (MusicFlare)", q.SenderID)
			if err != nil {
				_, _ = q.Answer("Unable to create playlist.")
				return nil
			}
		} else {
			playlistID = playlists[0].ID
		}

		song := db.Song{
			URL:      currentTrack.URL,
			Name:     currentTrack.Name,
			TrackID:  currentTrack.TrackID,
			Duration: currentTrack.Duration,
			Platform: currentTrack.Platform,
		}

		err = db.Instance.AddSongToPlaylist(playlistID, song)
		if err != nil {
			_, _ = q.Answer("Unable to add track to playlist.")
			return nil
		}

		playlist, err := db.Instance.GetPlaylist(playlistID)
		if err != nil {
			_, _ = q.Answer("Playlist not found.")
			return nil
		}

		_, _ = q.Answer(fmt.Sprintf("Track \"%s\" added to playlist \"%s\".", song.Name, playlist.Name))
		return nil
	}

	text := buildTrackMessage("Now Playing", "▶")
	_, _ = q.Edit(text, &tg.SendOptions{ReplyMarkup: core.ControlButtons("resume"), ParseMode: "HTML", LinkPreview: false})
	return nil
}

func vcPlayHandler(q *tg.CallbackQuery) error {
	data := q.DataString()

	if strings.Contains(data, "vcplay_close") {
		_, _ = q.Answer("Closing panel.")
		_, _ = client.DeleteMessages(q.ChatID, []int32{q.MessageID})
		return nil
	}

	slog.Info("Received vcplay callback", "arg1", data)
	return nil
}
