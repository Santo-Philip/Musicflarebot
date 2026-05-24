package handlers

import (
	"fmt"
	"musicflarebot/internal/ui"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func helpCommandHandler(m *tg.NewMessage) error {
	if !IsPrivate(m) {
		_, err := m.Reply("Click the button below for help.", &tg.SendOptions{
			ReplyMarkup: ui.SupportBtn(),
			ParseMode:   "HTML",
		})
		return err
	}

	response := getHelpText("all", firstName(m))
	_, err := m.Reply(response, &tg.SendOptions{
		ParseMode:   "HTML",
		LinkPreview: false,
		ReplyMarkup: ui.HelpMenuKeyboard(),
	})
	return err
}

func helpCallbackHandler(q *tg.CallbackQuery) error {
	data := q.DataString()
	if !strings.HasPrefix(data, "help_") {
		q.Answer("Processing...")
		return nil
	}

	cmd := strings.TrimPrefix(data, "help_")
	response := getHelpText(cmd, "")

	markup := ui.BackHelpMenuKeyboard()
	if cmd == "all" || cmd == "back" {
		markup = ui.HelpMenuKeyboard()
	}

	q.Answer("")
	_, err := q.Edit(response, &tg.SendOptions{
		ParseMode:   "HTML",
		LinkPreview: false,
		ReplyMarkup: markup,
	})
	return err
}

func getHelpText(cmd string, userName string) string {
	switch cmd {
	case "back", "all":
		return fmt.Sprintf(
			"<b>Help Menu</b>\n\n" +
				"ᴛʜᴇ ɪɴᴛᴇʟʟɪɢᴇɴᴛ ᴍᴜsɪᴄ ᴘʟᴀʏᴇʀ ʙᴏᴛ.\n\n" +
				"<b>ᴄʟɪᴄᴋ ᴏɴ ᴛʜᴇ ʙᴜᴛᴛᴏɴs ʙᴇʟᴏᴡ</b>\n\n" +
				"<b>ᴅᴇᴠᴇʟᴏᴘᴇʀ:</b> <a href='https://t.me/nexfang'>@NexFang</a>",
		)
	case "user":
		return `<b>User Commands</b>

<b>Available User Commands:</b>

 • <code>/play [song name or URL]</code> — Play a song
 • <code>/search [query]</code> — Search for a song
 • <code>/skip</code> — Skip to next track
 • <code>/pause</code> — Pause playback
 • <code>/resume</code> — Resume playback
 • <code>/stop</code> — Stop playback
 • <code>/mute</code> — Mute playback
 • <code>/unmute</code> — Unmute playback
 • <code>/queue</code> — View the queue
 • <code>/loop [count]</code> — Loop current track
 • <code>/remove [number]</code> — Remove a track from queue
 • <code>/seek [seconds]</code> — Seek within the current track
 • <code>/speed [0.5-4.0]</code> — Change playback speed

<b>Playlist Commands:</b>
 • <code>/createplaylist [name]</code> — Create a playlist
 • <code>/deleteplaylist [id]</code> — Delete a playlist
 • <code>/addtoplaylist [id] [url]</code> — Add song to playlist
 • <code>/removefromplaylist [id] [url]</code> — Remove song from playlist
 • <code>/playlistinfo [id]</code> — View playlist info
 • <code>/myplaylists</code> — View your playlists

<b>Note:</b> Use <code>/play</code> with a song name or YouTube/Spotify link.`
	case "admin":
		return `<b>Admin Commands</b>

<b>Available Admin Commands:</b>

 • <code>/settings</code> — Change chat settings
 • <code>/auth</code> — Authorize a user
 • <code>/addAuth</code> — Add authorized user
 • <code>/removeAuth</code> — Remove authorized user
 • <code>/authList</code> — List authorized users
 • <code>/reload</code> — Reload admin cache
 • <code>/privacy</code> — Show privacy policy

<b>Settings:</b>
 • <b>Play Mode:</b> Restrict /play to admins only
 • <b>Admin Mode:</b> Restrict all commands to admins
 • <b>Command Delete:</b> Auto-delete command messages
 • <b>Language:</b> Select bot language (coming soon)

<b>Note:</b> Use /settings in the group to change settings.`
	case "owner":
		return `<b>Owner Commands</b>

<b>Settings Commands:</b>
 • <code>/set_logger_id [id]</code> — Set logger channel ID
 • <code>/set_support_group [username]</code> — Set support group
 • <code>/set_support_channel [username]</code> — Set support channel
 • <code>/set_song_duration [seconds]</code> — Set max song duration
 • <code>/set_default_service [service]</code> — Set default streaming service
 • <code>/add_dev [id]</code> — Add a developer
 • <code>/remove_dev [id]</code> — Remove a developer
 • <code>/list_settings</code> — List current settings

<b>Account Commands:</b>
 • <code>/addacc</code> — Add an assistant account (interactive)
 • <code>/listacc</code> — List all configured assistant accounts
 • <code>/removeacc [index]</code> — Remove assistant account by index
 • <code>/cancel</code> — Cancel active /addacc session

<b>Note:</b> These commands are for bot owners only.`
	case "devs":
		return `<b>Developer Commands</b>

<b>Available Developer Commands:</b>

 • <code>/stats</code> — View system statistics
 • <code>/shell</code> or <code>/sh</code> — Execute shell commands
 • <code>/broadcast</code> or <code>/gcast</code> — Send broadcast to all chats
 • <code>/stop_broadcast</code> — Stop active broadcast
 • <code>/active_vc</code> or <code>/av</code> — Show active voice chats
 • <code>/clearAssistants</code> — Clear all assistant assignments
 • <code>/leaveAll</code> — Leave all chats
 • <code>/logger [on/off]</code> — Toggle logger status`
	case "playlist":
		return `<b>Playlist Commands</b>

<b>Available Playlist Commands:</b>

 • <code>/createplaylist [name]</code> — Create a new playlist
 • <code>/deleteplaylist [id]</code> — Delete a playlist
 • <code>/addtoplaylist [id] [url]</code> — Add a song to a playlist
 • <code>/removefromplaylist [id] [url]</code> — Remove a song from a playlist
 • <code>/playlistinfo [id]</code> — View playlist details
 • <code>/myplaylists</code> — View all your playlists

<b>Usage:</b>
 • Playlists allow you to save your favorite songs.
 • Maximum 10 playlists per user.
 • Maximum 50 songs per playlist.

<b>Note:</b> Use your playlist ID to manage songs.`
	default:
		return fmt.Sprintf("Unknown help section: %s", cmd)
	}
}
