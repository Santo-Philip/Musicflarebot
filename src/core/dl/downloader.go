package dl

import (
	"musicflarebot/internal/config"
	"musicflarebot/internal/database"
	"musicflarebot/internal/types"
	"musicflarebot/internal/utils"
	"fmt"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func DownloadCachedTrack(cached *types.CachedTrack, bot *tg.Client) (string, error) {
	if cached.Platform == types.DirectLink {
		return cached.URL, nil
	}

	if cached.Platform == types.Telegram {
		return downloadTelegramFile(cached, bot)
	}

	if cached.Platform == types.YouTube && cached.TrackID != "" {
		path, err := checkSongCache(cached.TrackID, bot)
		if err == nil && path != "" {
			return path, nil
		}
	}

	path, err := downloadViaWrapper(cached, bot)
	if err != nil {
		return "", err
	}

	if !cached.IsVideo && cached.Platform == types.YouTube && cached.TrackID != "" {
		_ = cacheSongFile(bot, cached.TrackID, path)
	}

	return path, nil
}

func downloadViaWrapper(cached *types.CachedTrack, bot *tg.Client) (string, error) {
	wrapper := NewDownloaderWrapper(cached.URL)
	if !wrapper.IsValid() {
		return "", fmt.Errorf("invalid cached URL: %s", cached.URL)
	}

	track, err := wrapper.GetTrack()
	if err != nil {
		return "", fmt.Errorf("get track info: %w", err)
	}

	path, err := wrapper.DownloadTrack(track, cached.IsVideo)
	if err != nil {
		return "", err
	}

	if utils.TelegramMessageRegex.MatchString(path) {
		return downloadFromTelegramMessage(bot, path)
	}

	return path, nil
}

func downloadTelegramFile(cached *types.CachedTrack, bot *tg.Client) (string, error) {
	// For Telegram platform, TrackID is a file ID. Use gogram's DownloadMedia.
	// Since we don't have a message object, we create one by sending a dummy request
	// or using raw API. For now, use GetMessageByID approach.
	// If TrackID looks like a t.me URL, resolve it.
	if utils.TelegramMessageRegex.MatchString(cached.TrackID) {
		return downloadFromTelegramMessage(bot, cached.TrackID)
	}
	return "", fmt.Errorf("cannot download telegram file with id: %s", cached.TrackID)
}

func downloadFromTelegramMessage(bot *tg.Client, msgURL string) (string, error) {
	msg, err := utils.GetMessage(bot, msgURL)
	if err != nil {
		return "", fmt.Errorf("get telegram message: %w", err)
	}

	path, err := msg.Download()
	if err != nil {
		return "", err
	}

	if path == "" {
		return "", fmt.Errorf("failed to download file from Telegram message")
	}

	return path, nil
}

func checkSongCache(trackID string, bot *tg.Client) (string, error) {
	loggerID := config.Conf.LoggerId
	if loggerID == 0 || trackID == "" {
		return "", nil
	}

	val, err := database.GetSetting("song_cache_" + trackID)
	if err != nil || val == "" {
		return "", nil
	}

	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 {
		return "", nil
	}

	msgID, err := strconv.Atoi(parts[1])
	if err != nil || msgID == 0 {
		return "", nil
	}

	url := fmt.Sprintf("https://t.me/c/%s/%d", parts[0], msgID)
	return downloadFromTelegramMessage(bot, url)
}

func cacheSongFile(bot *tg.Client, trackID, filePath string) error {
	loggerID := config.Conf.LoggerId
	if loggerID == 0 || trackID == "" || filePath == "" {
		return nil
	}

	msg, err := bot.SendMedia(loggerID, filePath, &tg.MediaOptions{
		ForceDocument: false,
	})
	if err != nil {
		return err
	}

	chatID := loggerID
	if chatID < 0 {
		s := strconv.FormatInt(-chatID, 10)
		chatStr := strings.TrimPrefix(s, "100")
		return database.SetSetting("song_cache_"+trackID, chatStr+":"+strconv.Itoa(int(msg.ID)))
	}

	return database.SetSetting("song_cache_"+trackID, strconv.FormatInt(chatID, 10)+":"+strconv.Itoa(int(msg.ID)))
}
