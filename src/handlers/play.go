package handlers

import (
	"fmt"
	"html"
	"log/slog"
	"musicflarebot/internal/config"
	"musicflarebot/internal/ui"
	"musicflarebot/internal/cache"
	"musicflarebot/internal/database"
	"musicflarebot/src/core/dl"
	"musicflarebot/src/vc"
	"strings"

	"musicflarebot/internal/types"
	"musicflarebot/internal/utils"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func playHandler(m *tg.NewMessage) error {
	if !playMode(m) {
		return nil
	}
	return handlePlay(m, false)
}

func vPlayHandler(m *tg.NewMessage) error {
	if !playMode(m) {
		return nil
	}
	return handlePlay(m, true)
}

func handlePlay(m *tg.NewMessage, isVideo bool) error {
	chatID := m.ChatID()
	slog.Debug("handlePlay called", "chatID", chatID, "isVideo", isVideo)

	if qLen := cache.ChatCache.GetQueueLength(chatID); qLen > 10 {
		_, _ = m.Reply("Queue is full (max 10 tracks). Use /end to clear.")
		return nil
	}

	isReply := m.ReplyToMsgID() != 0
	args := Args(m)
	url := getUrl(m, isReply)

	rMsg := m
	var err error
	if isReply && args == "" && url == "" {
		reply, err := getReplyMessage(m)
		if err == nil && reply != nil {
			args = reply.Text()
		}
	}

	input := coalesce(url, args)

	if strings.HasPrefix(input, "tgpl_") {
		playlist, err := database.GetPlaylist(input)
		if err != nil {
			_, err = m.Reply("❌ Playlist not found.")
			return err
		}

		tracks := database.ConvertSongsToTracks(playlist.Songs)
		if len(tracks) == 0 {
			_, err = m.Reply("❌ Playlist is empty.")
			return err
		}

		updater, err := m.Reply("🔍 Searching playlist...")
		if err != nil {
			slog.Warn("failed to send message", "error", err)
			return nil
		}

		return handleMultipleTracks(m, updater, tracks, chatID, isVideo)
	}

	if match := utils.TelegramMessageRegex.FindStringSubmatch(input); match != nil {
		rMsg, err = utils.GetMessage(client, input)
		if err != nil {
			slog.Warn("failed to parse message", "error", err.Error())
			_, err = m.Reply("Invalid Telegram link.")
			return err
		}
	} else if isReply {
		rMsg, err = getReplyMessage(m)
		if err != nil {
			_, err = m.Reply("Invalid reply message.")
			return err
		}
	}

	if valid := isValidMedia(rMsg); valid {
		isReply = true
	}

	if url == "" && args == "" && (!isReply || !isValidMedia(rMsg)) {
		_, _ = m.Reply("<b>Usage:</b>\n/play [song or URL]\n\n<b>Supported Platforms:</b>\n- YouTube\n- Spotify\n- JioSaavn\n- Apple Music", &tg.SendOptions{
			ReplyMarkup: ui.SupportKeyboard(),
			ParseMode:   "HTML",
		})
		return nil
	}

	updater, err := m.Reply("🔍 Searching and downloading...")
	if err != nil {
		slog.Warn("failed to send message", "error", err)
		return nil
	}

	if isReply && isValidMedia(rMsg) {
		return handleMedia(m, updater, rMsg, chatID, isVideo)
	}

	wrapper := dl.NewDownloaderWrapper(input)
	if url != "" {
		if !wrapper.IsValid() {
			_, _ = updater.Edit("Invalid URL or unsupported platform.\n\n<b>Supported Platforms:</b>\n- YouTube\n- Spotify\n- JioSaavn\n- Apple Music", &tg.SendOptions{
				ReplyMarkup: ui.SupportKeyboard(),
				ParseMode:   "HTML",
			})
			return nil
		}

		trackInfo, err := wrapper.GetInfo()
		if err != nil {
			_, _ = updater.Edit(fmt.Sprintf("❌ Error fetching track info: %s", err.Error()))
			return nil
		}

		if trackInfo.Results == nil || len(trackInfo.Results) == 0 {
			_, _ = updater.Edit("No tracks found.")
			return nil
		}

		return handleUrl(m, updater, trackInfo, chatID, isVideo)
	}

	return handleTextSearch(m, updater, wrapper, chatID, isVideo)
}

func handleMedia(m *tg.NewMessage, updater *tg.NewMessage, dlMsg *tg.NewMessage, chatID int64, isVideo bool) error {
	fileName, fileSize := getFileInfo(dlMsg)
	if fileName == "" {
		_, err := updater.Edit("No valid media found in the message.")
		return err
	}

	if fileSize > config.Conf.MaxFileSize {
		_, err := updater.Edit(fmt.Sprintf("File too large. Max size: %d MB.", config.Conf.MaxFileSize/(1024*1024)))
		if err != nil {
			slog.Warn("Edit message failed", "error", err)
		}
		return nil
	}

	fileID := fmt.Sprintf("%d_%d", dlMsg.ChatID(), dlMsg.ID)
	if _track := cache.ChatCache.GetTrackIfExists(chatID, fileID); _track != nil {
		_, err := updater.Edit("Track already in queue or playing.")
		return err
	}

	dur := utils.GetFileDur(dlMsg)
	link := getMessageLink(dlMsg)

	saveCache := types.CachedTrack{
		URL: link, Name: fileName, User: firstName(m), UserID: m.SenderID(), TrackID: fileID,
		Duration: dur, IsVideo: isVideo, Platform: types.Telegram,
	}

	qLen := cache.ChatCache.AddSong(chatID, &saveCache)
	if qLen > 1 {
		escURL := html.EscapeString(saveCache.URL)
		escName := html.EscapeString(saveCache.Name)
		escUser := html.EscapeString(saveCache.User)
		queueInfo := fmt.Sprintf(
			"<u><b>Added to queue: %d</b></u>\n\n<b>Title:</b> <a href='%s'>%s</a>\n\n<b>Duration:</b> %s min\n<b>Requested by:</b> %s",
			qLen, escURL, escName, utils.SecToMin(saveCache.Duration), escUser,
		)
		_, err := updater.Edit(queueInfo, &tg.SendOptions{ReplyMarkup: ui.ControlButtons("play"), ParseMode: "HTML", LinkPreview: false})
		return err
	}

	filePath, err := dlMsg.Download()
	if err != nil {
		cache.ChatCache.RemoveCurrentSong(chatID)
		_, err = updater.Edit(fmt.Sprintf("Download failed: %s", err.Error()))
		return err
	}

	if dur == 0 {
		dur = utils.GetMediaDuration(filePath)
		saveCache.Duration = dur
	}

	saveCache.FilePath = filePath

	if err = vc.Calls.PlayMedia(chatID, saveCache.FilePath, saveCache.IsVideo, ""); err != nil {
		cache.ChatCache.RemoveCurrentSong(chatID)
		_, err = updater.Edit(html.EscapeString(err.Error()), &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
		return err
	}

	escURL := html.EscapeString(saveCache.URL)
	escName := html.EscapeString(saveCache.Name)
	escUser := html.EscapeString(saveCache.User)

	nowPlaying := fmt.Sprintf(
		"<u><b>| Started streaming</b></u>\n\n<b>Title:</b> <a href='%s'>%s</a>\n\n<b>Duration:</b> %s min\n<b>Requested by:</b> %s",
		escURL, escName, utils.SecToMin(saveCache.Duration), escUser,
	)

	_, err = updater.Edit(nowPlaying, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: ui.ControlButtons("play"),
		LinkPreview: false,
	})

	return err
}

func handleTextSearch(m *tg.NewMessage, updater *tg.NewMessage, wrapper *dl.DownloaderWrapper, chatID int64, isVideo bool) error {
	searchResult, err := wrapper.Search()
	if err != nil {
		_, err = updater.Edit(fmt.Sprintf("❌ Search failed: %s", err.Error()))
		return err
	}

	if searchResult.Results == nil || len(searchResult.Results) == 0 {
		_, err = updater.Edit("😕 No results found. Try a different query.")
		return err
	}

	song := searchResult.Results[0]
	if _track := cache.ChatCache.GetTrackIfExists(chatID, song.Id); _track != nil {
		_, err := updater.Edit("Track already in queue or playing.")
		return err
	}

	return handleSingleTrack(m, updater, song, "", chatID, isVideo)
}

func handleUrl(m *tg.NewMessage, updater *tg.NewMessage, trackInfo types.PlatformTracks, chatID int64, isVideo bool) error {
	if len(trackInfo.Results) == 1 {
		track := trackInfo.Results[0]
		if _track := cache.ChatCache.GetTrackIfExists(chatID, track.Id); _track != nil {
			_, err := updater.Edit("Track already in queue or playing.")
			return err
		}
		return handleSingleTrack(m, updater, track, "", chatID, isVideo)
	}

	return handleMultipleTracks(m, updater, trackInfo.Results, chatID, isVideo)
}

func handleSingleTrack(m *tg.NewMessage, updater *tg.NewMessage, song types.MusicTrack, filePath string, chatID int64, isVideo bool) error {
	if song.Duration > int(config.Conf.SongDurationLimit) {
		_, err := updater.Edit(fmt.Sprintf("Sorry, song exceeds max duration of %d minutes.", config.Conf.SongDurationLimit/60))
		return err
	}

	saveCache := types.CachedTrack{
		URL: song.Url, Name: song.Title, User: firstName(m), UserID: m.SenderID(), FilePath: filePath,
		Thumbnail: song.Thumbnail, TrackID: song.Id, Duration: song.Duration, Channel: song.Channel, Views: song.Views,
		IsVideo: isVideo, Platform: song.Platform,
	}

	qLen := cache.ChatCache.AddSong(chatID, &saveCache)
	if qLen > 1 {
		escURL := html.EscapeString(saveCache.URL)
		escName := html.EscapeString(saveCache.Name)
		escUser := html.EscapeString(saveCache.User)
		queueInfo := fmt.Sprintf(
			"<u><b>Added to queue: %d</b></u>\n\n<b>Title:</b> <a href='%s'>%s</a>\n\n<b>Duration:</b> %s min\n<b>Requested by:</b> %s",
			qLen, escURL, escName, utils.SecToMin(saveCache.Duration), escUser,
		)

		_, err := updater.Edit(queueInfo, &tg.SendOptions{ReplyMarkup: ui.ControlButtons("play"), ParseMode: "HTML", LinkPreview: false})
		return err
	}

	if saveCache.FilePath == "" {
		dlResult, err := dl.DownloadCachedTrack(&saveCache, client)
		if err != nil {
			cache.ChatCache.RemoveCurrentSong(chatID)
			_, err = updater.Edit(fmt.Sprintf("Download failed: %s", err.Error()))
			return err
		}

		saveCache.FilePath = dlResult
	}

	if err := vc.Calls.PlayMedia(chatID, saveCache.FilePath, saveCache.IsVideo, ""); err != nil {
		cache.ChatCache.RemoveCurrentSong(chatID)
		_, err = updater.Edit(html.EscapeString(err.Error()), &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
		return err
	}

	escURLnp := html.EscapeString(saveCache.URL)
	escNamenp := html.EscapeString(saveCache.Name)
	escUsernp := html.EscapeString(saveCache.User)

	nowPlaying := fmt.Sprintf(
		"<u><b>| Started streaming</b></u>\n\n<b>Title:</b> <a href='%s'>%s</a>\n\n<b>Duration:</b> %s min\n<b>Requested by:</b> %s",
		escURLnp, escNamenp, utils.SecToMin(song.Duration), escUsernp,
	)

	_, err := updater.Edit(nowPlaying, &tg.SendOptions{
		ReplyMarkup: ui.ControlButtons("play"),
		ParseMode:   "HTML",
		LinkPreview: false,
	})

	if err != nil {
		slog.Warn("Edit message failed", "error", err)
		return err
	}

	return nil
}

func handleMultipleTracks(m *tg.NewMessage, updater *tg.NewMessage, tracks []types.MusicTrack, chatID int64, isVideo bool) error {
	if len(tracks) == 0 {
		_, err := updater.Edit("No tracks found.")
		return err
	}

	queueHeader := "<u><b>Added to Queue:</b></u>\n<blockquote expandable>\n"
	var tracksToAdd []*types.CachedTrack
	var skippedTracks []string

	shouldPlayFirst := false
	var firstTrack *types.CachedTrack

	for _, track := range tracks {
		if track.Duration > int(config.Conf.SongDurationLimit) {
			skippedTracks = append(skippedTracks, track.Title)
			continue
		}

		saveCache := &types.CachedTrack{
			Name: track.Title, TrackID: track.Id, Duration: track.Duration,
			Thumbnail: track.Thumbnail, User: firstName(m), UserID: m.SenderID(), Platform: track.Platform,
			IsVideo: isVideo, URL: track.Url, Channel: track.Channel, Views: track.Views,
		}
		tracksToAdd = append(tracksToAdd, saveCache)
	}

	if len(tracksToAdd) == 0 {
		if len(skippedTracks) > 0 {
			_, err := updater.Edit(fmt.Sprintf("All tracks were skipped (max duration %d min).", config.Conf.SongDurationLimit/60))
			return err
		}
		_, err := updater.Edit("No valid tracks found.")
		return err
	}

	qLenAfter := cache.ChatCache.AddSongs(chatID, tracksToAdd)
	startLen := qLenAfter - len(tracksToAdd)

	if startLen == 0 {
		shouldPlayFirst = true
		firstTrack = tracksToAdd[0]
		firstTrack.Loop = 1
	}

	var sb strings.Builder
	sb.WriteString(queueHeader)

	totalDuration := 0
	for i, track := range tracksToAdd {
		currentQLen := startLen + i + 1
		escTrackName := html.EscapeString(track.Name)
		fmt.Fprintf(&sb, "<b>%d.</b> %s\n└ Duration: %s\n",
			currentQLen, escTrackName, utils.SecToMin(track.Duration))
		totalDuration += track.Duration
	}

	sb.WriteString("</blockquote>")
	escRequester := html.EscapeString(firstName(m))
	queueSummary := fmt.Sprintf(
		"\n<b>Queue Total:</b> %d\n<b>Duration:</b> %s min\n<b>Requested by:</b> %s",
		qLenAfter, utils.SecToMin(totalDuration), escRequester,
	)

	sb.WriteString(queueSummary)
	if len(skippedTracks) > 0 {
		fmt.Fprintf(&sb, "\n\n<b>Skipped %d tracks</b> (exceeded duration limit).", len(skippedTracks))
	}

	fullMessage := sb.String()

	if len(fullMessage) > 4096 {
		fullMessage = queueSummary
	}

	if shouldPlayFirst && firstTrack != nil {
		_ = vc.Calls.PlayNext(chatID)
	}

	_, err := updater.Edit(fullMessage, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: ui.ControlButtons("play"),
		LinkPreview: false,
	})

	return err
}

func getMessageLink(m *tg.NewMessage) string {
	chatID := m.ChatID()
	msgID := m.ID
	if chatID > 0 {
		return ""
	}
	s := fmt.Sprintf("%d", chatID)
	s = strings.TrimPrefix(s, "-100")
	return fmt.Sprintf("https://t.me/c/%s/%d", s, msgID)
}
