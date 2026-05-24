package types

const (
	Telegram   = "telegram"
	YouTube    = "youtube"
	Spotify    = "spotify"
	JioSaavn   = "jiosaavn"
	Apple      = "apple_music"
	SoundCloud = "soundcloud"
	Deezer     = "Deezer"
	Gaana      = "Gaana"
	DirectLink = "direct_link"
	Tidal      = "tidal"
	MXPlayer   = "mxplayer"
	Twitch     = "twitch"
	TwitchClip = "twitch_clip"
	Kick       = "kick"
	KickClip   = "kick_clip"
)

const (
	Admins   = "admins"
	Everyone = "everyone"
)

const (
	PlatformYouTube    = "youtube"
	PlatformSpotify    = "spotify"
	PlatformJioSaavn   = "jiosaavn"
	PlatformAppleMusic = "applemusic"
	PlatformSoundCloud = "soundcloud"
)

var ValidServices = []string{
	PlatformYouTube,
	PlatformSpotify,
	PlatformJioSaavn,
	PlatformAppleMusic,
	PlatformSoundCloud,
}

const (
	SettingLoggerId       = "logger_id"
	SettingSupportGroup   = "support_group"
	SettingSupportChannel = "support_channel"
	SettingSongDuration   = "song_duration_limit"
	SettingDefaultService = "default_service"
	SettingDevs           = "devs"
	SettingStartMessage   = "start_message"
	SettingStartMedia     = "start_media"
	SettingStartMediaType = "start_media_type"
)
