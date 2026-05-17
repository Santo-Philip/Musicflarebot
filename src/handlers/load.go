package handlers

import (
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var startTime = time.Now()
var client *tg.Client

func LoadModules(c *tg.Client) {
	client = c

	c.OnCommand("reload", reloadAdminCacheHandler)
	c.OnCommand("authList", authListHandler)
	c.OnCommand("auths", authListHandler)
	c.OnCommand("auth", addAuthHandler)
	c.OnCommand("addAuth", addAuthHandler)
	c.OnCommand("removeAuth", removeAuthHandler)
	c.OnCommand("rmAuth", removeAuthHandler)
	c.OnCommand("broadcast", broadcastHandler)
	c.OnCommand("gCast", broadcastHandler)
	c.OnCommand("stop_gcast", cancelBroadcastHandler)
	c.OnCommand("stop_broadcast", cancelBroadcastHandler)
	c.OnCommand("av", activeVcHandler)
	c.OnCommand("active_vc", activeVcHandler)
	c.OnCommand("clearass", clearAssistantsHandler)
	c.OnCommand("clearAssistants", clearAssistantsHandler)
	c.OnCommand("leaveAll", leaveAllHandler)
	c.OnCommand("logger", loggerHandler)
	c.OnCommand("privacy", privacyHandler)
	c.OnCommand("loop", loopHandler)
	c.OnCommand("pause", pauseHandler)
	c.OnCommand("resume", resumeHandler)
	c.OnCommand("cplist", createPlaylistHandler)
	c.OnCommand("createplaylist", createPlaylistHandler)
	c.OnCommand("deleteplaylist", deletePlaylistHandler)
	c.OnCommand("queue", queueHandler)
	c.OnCommand("seek", seekHandler)
	c.OnCommand("sh", shellCommand)
	c.OnCommand("skip", skipHandler)
	c.OnCommand("speed", speedHandler)
	c.OnCommand("stop", stopHandler)
	c.OnCommand("end", stopHandler)
	c.OnCommand("start", startHandler)
	c.OnCommand("help", helpCommandHandler)
	c.OnCommand("ping", pingHandler)
	c.OnCommand("play", playHandler)
	c.OnCommand("p", playHandler)
	c.OnCommand("vplay", vPlayHandler)
	c.OnCommand("v", vPlayHandler)
	c.OnCommand("remove", removeHandler)
	c.OnCommand("mute", muteHandler)
	c.OnCommand("unmute", unmuteHandler)
	c.OnCommand("settings", settingsHandler)
	c.OnCommand("addtoplaylist", addToPlaylistHandler)
	c.OnCommand("addtoplist", addToPlaylistHandler)
	c.OnCommand("removefromplaylist", removeFromPlaylistHandler)
	c.OnCommand("rmplist", removeFromPlaylistHandler)
	c.OnCommand("plistinfo", playlistInfoHandler)
	c.OnCommand("playlistinfo", playlistInfoHandler)
	c.OnCommand("myplaylists", myPlaylistsHandler)
	c.OnCommand("myplist", myPlaylistsHandler)
	c.OnCommand("stats", statsHandler)

	c.OnCommand("set_logger_id", setLoggerIdHandler)
	c.OnCommand("set_support_group", setSupportGroupHandler)
	c.OnCommand("set_support_channel", setSupportChannelHandler)
	c.OnCommand("set_song_duration", setSongDurationHandler)
	c.OnCommand("set_default_service", setDefaultServiceHandler)
	c.OnCommand("add_dev", addDevHandler)
	c.OnCommand("remove_dev", removeDevHandler)
	c.OnCommand("list_settings", listSettingsHandler)

	c.OnCommand("addacc", addaccHandler)
	c.OnCommand("cancel", cancelAddAccHandler)
	c.OnCommand("listacc", listAccHandler)
	c.OnCommand("removeacc", removeAccHandler)
	c.OnCommand("setcookies", setCookiesHandler)

	c.OnCallback("", func(q *tg.CallbackQuery) error {
		data := q.DataString()
		switch {
		case data == "help_back":
			return helpCallbackHandler(q)
		case data == "vcplay_close":
			return vcPlayHandler(q)
		case data == "help_all" || data == "help_user" || data == "help_admin" || data == "help_owner" || data == "help_devs" || data == "help_playlist":
			return helpCallbackHandler(q)
		case data == "settings_main" || data == "settings_delete" || data == "settings_play" || data == "settings_admin" || data == "settings_lang":
			return settingsCallbackHandler(q)
		default:
			if len(data) > 5 && data[:5] == "play_" {
				return playCallbackHandler(q)
			}
			q.Answer("Unknown action")
			return nil
		}
	})

	c.OnMessage("", func(m *tg.NewMessage) error {
		if err := handleVoiceChatMessage(m); err != nil {
			return err
		}
		if err := handleAddAccMessage(m); err != nil {
			return err
		}
		return nil
	})

	c.OnParticipant(handleParticipant)
}
