package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"musicflarebot/internal/database"
	"musicflarebot/src/core/dl"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func createPlaylistHandler(m *tg.NewMessage) error {
	userID := m.SenderID()

	args := Args(m)
	if args == "" {
		_, err := m.Reply("<b>Usage:</b> /createplaylist [playlist name]", replyOpts)
		return err
	}

	userPlaylists, err := database.GetUserPlaylists(userID)
	if err != nil {
		_, err = m.Reply("Unable to fetch your playlists. Please try again later.")
		return err
	}

	if len(userPlaylists) >= 10 {
		_, _ = m.Reply("You have reached the maximum limit of 10 playlists.")
		return nil
	}

	if len([]rune(args)) > 40 {
		args = string([]rune(args)[:40])
	}

	playlistID, err := database.CreatePlaylist(args, userID)
	if err != nil {
		_, err = m.Reply(fmt.Sprintf("Failed to create playlist: %s", err.Error()))
		return err
	}

	_, err = m.Reply(
		fmt.Sprintf(
			"Playlist <b>%s</b> has been created successfully.\nID: <code>%s</code>",
			args,
			playlistID,
		),
		replyOpts,
	)

	return nil
}

func deletePlaylistHandler(m *tg.NewMessage) error {
	userID := m.SenderID()

	args := Args(m)
	if args == "" {
		_, err := m.Reply(
			"<b>Usage:</b> /deleteplaylist [playlist id]",
			&tg.SendOptions{ParseMode: "HTML"},
		)
		return err
	}

	playlist, err := database.GetPlaylist(args)
	if err != nil {
		_, err := m.Reply(
			"The specified playlist could not be found. Please check the playlist ID.",
		)
		return err
	}

	if playlist.UserID != userID {
		_, err := m.Reply(
			"You can only delete playlists that you created.",
		)
		return err
	}

	err = database.DeletePlaylist(args, userID)
	if err != nil {
		_, err := m.Reply(
			fmt.Sprintf("Failed to delete the playlist: %s", err.Error()),
		)
		return err
	}

	_, err = m.Reply(
		fmt.Sprintf("Playlist <b>%s</b> has been deleted successfully.", playlist.Name),
		&tg.SendOptions{ParseMode: "HTML"},
	)

	return err
}

func addToPlaylistHandler(m *tg.NewMessage) error {
	userID := m.SenderID()

	args := strings.SplitN(Args(m), " ", 2)
	if len(args) != 2 {
		_, err := m.Reply(
			"<b>Usage:</b> /addtoplaylist [playlist id] [song url]",
			&tg.SendOptions{ParseMode: "HTML"},
		)
		return err
	}

	playlistID := args[0]
	songURL := args[1]

	playlist, err := database.GetPlaylist(playlistID)
	if err != nil {
		_, err := m.Reply(
			"The specified playlist could not be found. Please verify the playlist ID.",
		)
		return err
	}

	if playlist.UserID != userID {
		_, err := m.Reply(
			"You can only modify playlists that you created.",
		)
		return err
	}

	wrapper := dl.NewDownloaderWrapper(songURL)
	if !wrapper.IsValid() {
		_, err := m.Reply(
			"The provided URL is invalid or the platform is not supported.",
		)
		return err
	}

	trackInfo, err := wrapper.GetInfo()
	if err != nil {
		_, err := m.Reply(
			fmt.Sprintf("Unable to retrieve track information: %s", err.Error()),
		)
		return err
	}

	if trackInfo.Results == nil || len(trackInfo.Results) == 0 {
		_, err := m.Reply(
			"No playable tracks were found for the provided link.",
		)
		return err
	}

	song := database.Song{
		URL:      trackInfo.Results[0].Url,
		Name:     trackInfo.Results[0].Title,
		TrackID:  trackInfo.Results[0].Id,
		Duration: trackInfo.Results[0].Duration,
		Platform: trackInfo.Results[0].Platform,
	}

	err = database.AddSongToPlaylist(playlistID, song)
	if err != nil {
		_, err := m.Reply(
			fmt.Sprintf("Failed to add the track to the playlist: %s", err.Error()),
		)
		return err
	}

	_, err = m.Reply(
		fmt.Sprintf(
			"Track <b>%s</b> has been added to playlist <b>%s</b>.",
			song.Name,
			playlist.Name,
		),
		replyOpts,
	)

	return err
}

func removeFromPlaylistHandler(m *tg.NewMessage) error {
	userID := m.SenderID()

	args := strings.SplitN(Args(m), " ", 2)
	if len(args) != 2 {
		_, err := m.Reply(
			"<b>Usage:</b> /removefromplaylist [playlist id] [song number or url]",
			&tg.SendOptions{ParseMode: "HTML"},
		)
		return err
	}

	playlistID := args[0]
	songIdentifier := args[1]

	playlist, err := database.GetPlaylist(playlistID)
	if err != nil {
		_, err = m.Reply("Playlist not found.")
		return err
	}

	if playlist.UserID != userID {
		_, err = m.Reply("You do not own this playlist.")
		return err
	}

	songIndex, err := strconv.Atoi(songIdentifier)
	var trackID string

	if err == nil {
		if songIndex < 1 || songIndex > len(playlist.Songs) {
			_, err := m.Reply("Invalid song number.")
			return err
		}
		trackID = playlist.Songs[songIndex-1].TrackID
	} else {
		for _, song := range playlist.Songs {
			if song.URL == songIdentifier || song.TrackID == songIdentifier {
				trackID = song.TrackID
				break
			}
		}
	}

	if trackID == "" {
		_, err = m.Reply("Song not found in playlist.")
		return err
	}

	err = database.RemoveSongFromPlaylist(playlistID, trackID)
	if err != nil {
		_, err = m.Reply(fmt.Sprintf("Error removing song: %s", err.Error()))
		return err
	}

	_, err = m.Reply(fmt.Sprintf("Song removed from playlist '%s'.", playlist.Name))
	return err
}

func playlistInfoHandler(m *tg.NewMessage) error {
	args := Args(m)
	if args == "" {
		_, err := m.Reply(
			"<b>Usage:</b> /playlistinfo [playlist id]",
			&tg.SendOptions{ParseMode: "HTML"},
		)
		return err
	}

	playlist, err := database.GetPlaylist(args)
	if err != nil {
		_, err = m.Reply("Playlist not found.")
		return err
	}

	var songs []string
	for i, song := range playlist.Songs {
		songs = append(songs, fmt.Sprintf("%d. %s (%s)", i+1, song.Name, song.URL))
	}

	owner, err := client.GetUser(playlist.UserID)
	if err != nil {
		return nil
	}

	_, err = m.Reply(
		fmt.Sprintf(
			"<b>Playlist Info</b>\n\n<b>Name:</b> %s\n<b>Owner:</b> %s\n<b>Songs:</b> %d\n\n%s",
			playlist.Name,
			owner.FirstName,
			len(playlist.Songs),
			strings.Join(songs, "\n"),
		),
		&tg.SendOptions{ParseMode: "HTML"},
	)
	return nil
}

func myPlaylistsHandler(m *tg.NewMessage) error {
	userID := m.SenderID()

	playlists, err := database.GetUserPlaylists(userID)
	if err != nil {
		_, err := m.Reply(fmt.Sprintf("Error fetching playlists: %s", err.Error()))
		return err
	}

	if len(playlists) == 0 {
		_, err := m.Reply("You do not have any playlists.")
		return err
	}

	var playlistInfo []string
	for _, playlist := range playlists {
		playlistInfo = append(
			playlistInfo,
			fmt.Sprintf("- %s (<code>%s</code>)", playlist.Name, playlist.ID),
		)
	}

	_, err = m.Reply(
		fmt.Sprintf("<b>My Playlists</b>\n\n%s", strings.Join(playlistInfo, "\n")),
		&tg.SendOptions{ParseMode: "HTML"},
	)

	return err
}
