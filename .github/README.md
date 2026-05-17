<h1 align="center">🎵 MusicFlare Bot (Go)</h1>

<p align="center">
  <a href="https://golang.org/">
    <img src="https://img.shields.io/badge/Written%20in-Go-blue?style=for-the-badge&logo=go" alt="Written in Go">
  </a>
  <a href="https://github.com/FlareBase/MusicFlareBot/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-GPL%20v3-green?style=for-the-badge" alt="License">
  </a>
  <a href="https://github.com/FlareBase/MusicFlareBot/stargazers">
    <img src="https://img.shields.io/github/stars/FlareBase/MusicFlareBot?style=for-the-badge&color=ffd700&logo=github" alt="Stars">
  </a>
</p>

<p align="center">
  A high-performance, feature-rich Telegram Music Bot written in <b>Go</b>. <br>
  Built with <code>gogram</code> (MTProto), <code>ntgcalls</code>, and <code>PostgreSQL</code>.<br>
  Developer: <a href="https://t.me/nexfang">@nexfang</a>
</p>

---

## Features

- **High Performance**: Written in Go for efficiency and speed.
- **Video & Audio**: Supports playing both video and audio streams.
- **Multi-Source**: YouTube, Direct Links, M3U8, etc.
- **Advanced Control**: Seek, Pause, Resume, Mute, Volume Control.
- **Playlist Management**: Queue system with skip, loop, and shuffle.
- **Multi-Language**: Easy to localize for different regions.
- **Unlimited Assistant Accounts**: Add/remove accounts interactively via `/addacc`.
- **PostgreSQL**: Persistent storage for settings, playlists, and auth.
- **Logger Channel**: Song caching via logger channel for faster playback.

---

## Installation & Setup

### Prerequisites
- **Go 1.22+**
- **FFmpeg**
- **ntgcalls C library** (`pip install ntgcalls`)
- **PostgreSQL** (or use a hosted service like Neon, Supabase, etc.)

### Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/FlareBase/MusicFlareBot.git
   cd MusicFlareBot
   ```

2. **Setup ntgcalls**
   ```bash
   pip install ntgcalls
   ```

3. **Configure Environment**
   ```bash
   cp sample.env .env
   # Edit .env with your credentials
   ```

4. **Build and Run**
   ```bash
   go build -o musicflarebot .
   ./musicflarebot
   ```

5. **Run in Background**
   - **Quick Start**: Use `screen` or `tmux`.
   - **Production**: Use `systemd`.

---

## Configuration

The bot is configured via environment variables. See `sample.env` for all options.

| Variable       | Description                               | Required |
|:---------------|:------------------------------------------|:--------:|
| `API_ID`       | Telegram API ID                           |    Yes    |
| `API_HASH`     | Telegram API Hash                         |    Yes    |
| `TOKEN`        | Bot Token from @BotFather                 |    Yes    |
| `DATABASE_URL` | PostgreSQL Connection URI                 |    Yes    |
| `API_URL`      | Your API URL (for song downloads)         |    Yes    |
| `API_KEY`      | Your API key                              |    Yes    |
| `OWNER_ID`     | Telegram User ID of the owner             |    Yes    |
| `SESSION_TYPE` | Session type (default: `session_db`)      |    No    |

Runtime settings (configured via bot commands, persisted in DB):
- Logger Channel ID
- Song Duration Limit
- Default Service
- Support Group / Channel
- Dev Users

---

## Commands

### Admin Commands
- `/play <query|url>` - Play a song or video.
- `/vplay <query|url>` - Force video play.
- `/skip` - Skip the current track.
- `/pause` - Pause playback.
- `/resume` - Resume playback.
- `/end` - Stop playback and clear queue.
- `/mute` - Mute the assistant.
- `/unmute` - Unmute the assistant.
- `/auth` - Authorize a user to use the bot.
- `/removeAuth` - Revoke authorization.
- `/settings` - Configure bot settings.
- `/loop` - Toggle loop mode.
- `/seek` - Seek to a position.
- `/speed` - Change playback speed.
- `/queue` - View the current queue.

### Playlist Commands
- `/createplaylist` - Create a new playlist.
- `/deleteplaylist` - Delete a playlist.
- `/addtoplaylist` - Add a song to a playlist.
- `/removefromplaylist` - Remove a song from a playlist.
- `/playlistinfo` - View playlist details.
- `/myplaylists` - List your playlists.

### Owner Commands
- `/addacc` - Add a new assistant account interactively.
- `/listacc` - List all assistant accounts.
- `/removeacc` - Remove an assistant account.
- `/add_dev` - Add a developer user.
- `/remove_dev` - Remove a developer user.
- `/list_settings` - View all runtime settings.
- `/set_logger_id` - Set the logger channel.
- `/set_song_duration` - Set max song duration.
- `/set_default_service` - Set default streaming service.
- `/set_support_group` - Set support group.
- `/set_support_channel` - Set support channel.
- `/broadcast` - Broadcast a message to all chats.
- `/stats` - View bot statistics.

### User Commands
- `/start` - Check if bot is alive.
- `/ping` - Check latency.
- `/help` - Show help menu.

---

## Links

- Repo: [MusicFlareBot on GitHub](https://github.com/Santo-Philip/Musicflarebot)
- Old version: TgMusicBot (Python)

---

<p align="center">
  Made with ❤️ by <a href="https://t.me/nexfang">nexfang</a><br>
  music flare bot by @Flarebase
</p>
