# MusicFlareBot - Features & Commands Reference

A comprehensive Telegram music player bot with multi-platform support, advanced controls, and group management features.

---

## 📋 Table of Contents

1. [User Commands](#user-commands)
2. [Admin Commands](#admin-commands)
3. [Owner Commands](#owner-commands)
4. [Developer Commands](#developer-commands)
5. [Playlist Management](#playlist-management)
6. [Supported Music Platforms](#supported-music-platforms)
7. [Feature Highlights](#feature-highlights)

---

## 🎵 User Commands

User commands available to all group members.

### Playback Control

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Play | `/play [song name or URL]` | Play a song by name or URL | `/play Bohemian Rhapsody` |
| Search | `/search [query]` | Search for a song | `/search 80s hits` |
| Skip | `/skip` | Skip to the next track | `/skip` |
| Pause | `/pause` | Pause current playback | `/pause` |
| Resume | `/resume` | Resume paused playback | `/resume` |
| Stop | `/stop` | Stop all playback and leave voice chat | `/stop` |

### Volume & Effects

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Mute | `/mute` | Mute audio playback | `/mute` |
| Unmute | `/unmute` | Unmute audio playback | `/unmute` |
| Speed | `/speed [0.5-4.0]` | Adjust playback speed | `/speed 1.5` |

### Queue Management

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Queue | `/queue` | View current queue | `/queue` |
| Loop | `/loop [count]` | Loop current track | `/loop 3` |
| Remove | `/remove [number]` | Remove track from queue | `/remove 2` |
| Seek | `/seek [seconds]` | Seek within current track | `/seek 120` |

### Playlist Commands

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Create Playlist | `/createplaylist [name]` | Create a new playlist | `/createplaylist MyMix` |
| Delete Playlist | `/deleteplaylist [id]` | Delete a playlist | `/deleteplaylist 1` |
| Add to Playlist | `/addtoplaylist [id] [url]` | Add song to playlist | `/addtoplaylist 1 https://youtube.com/watch?v=...` |
| Remove from Playlist | `/removefromplaylist [id] [url]` | Remove song from playlist | `/removefromplaylist 1 https://youtube.com/watch?v=...` |
| Playlist Info | `/playlistinfo [id]` | View playlist details | `/playlistinfo 1` |
| My Playlists | `/myplaylists` | View your playlists | `/myplaylists` |

### General

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Help | `/help` | Show help menu | `/help` |
| Start | `/start` | Start the bot | `/start` |

---

## 👨‍💼 Admin Commands

Commands available to group administrators.

### Group Settings

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Settings | `/settings` | Open settings menu | `/settings` |
| Play Mode | [via settings] | Restrict `/play` to admins only | - |
| Admin Mode | [via settings] | Restrict all commands to admins | - |
| Command Delete | [via settings] | Auto-delete command messages | - |
| Language | [via settings] | Select bot language (coming soon) | - |

### User Authorization

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Add Auth | `/addauth [user_id or @username]` | Authorize a user | `/addauth @john_doe` |
| Remove Auth | `/removeauth [user_id or @username]` | Revoke authorization | `/removeauth @john_doe` |
| Auth List | `/authlist` | List authorized users | `/authlist` |
| Auth | `/auth` | Manual authentication | `/auth` |

### Cache Management

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Reload Cache | `/reload` | Reload admin cache | `/reload` |

### Information

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Privacy | `/privacy` | Show privacy policy | `/privacy` |

---

## 👑 Owner Commands

Commands exclusive to the bot owner.

### Owner Settings

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Set Logger | `/set_logger_id [channel_id]` | Set bot log channel | `/set_logger_id -1001234567890` |
| Set Support Group | `/set_support_group [username]` | Set support group | `/set_support_group @botSupport` |
| Set Support Channel | `/set_support_channel [channel_id]` | Set support channel | `/set_support_channel @updates` |
| Set Song Duration | `/set_song_duration [seconds]` | Set max song duration | `/set_song_duration 600` |
| Set Default Service | `/set_default_service [service]` | Set default music service | `/set_default_service spotify` |
| List Settings | `/list_settings` | Display all owner settings | `/list_settings` |

### Developer Management

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Add Developer | `/add_dev [user_id]` | Add a developer | `/add_dev 123456789` |
| Remove Developer | `/remove_dev [user_id]` | Remove a developer | `/remove_dev 123456789` |

### Account Management

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Add Account | `/addacc` | Add assistant account (interactive) | `/addacc` |
| List Accounts | `/listacc` | List all configured accounts | `/listacc` |
| Remove Account | `/removeacc [index]` | Remove account by index | `/removeacc 1` |
| Cancel Account | `/cancel` | Cancel active `/addacc` session | `/cancel` |

---

## 🔧 Developer Commands

Advanced commands for bot developers and maintainers.

### System Information

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Stats | `/stats` | View system statistics | `/stats` |

### System Control

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Shell | `/shell [command]` or `/sh [command]` | Execute shell commands | `/shell ls -la` |
| Leave All | `/leaveAll` | Leave all voice chats | `/leaveAll` |
| Active Voice Chats | `/active_vc` or `/av` | Show active voice chats | `/active_vc` |

### Broadcast

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Broadcast | `/broadcast [message]` or `/gcast [message]` | Send to all chats | `/broadcast Maintenance scheduled` |
| Stop Broadcast | `/stop_broadcast` | Cancel active broadcast | `/stop_broadcast` |

### Maintenance

| Command | Syntax | Description | Example |
|---------|--------|-------------|---------|
| Clear Assistants | `/clearAssistants` | Clear assistant assignments | `/clearAssistants` |
| Logger Toggle | `/logger [on/off]` | Enable/disable logging | `/logger on` |

---

## 📚 Playlist Management

Detailed playlist functionality for organizing songs.

### Creating & Managing Playlists

**Create Playlist**
```
/createplaylist [name]
```
- Create a personal playlist
- Names can contain spaces
- Stored in database (persistent)

**View Playlists**
```
/myplaylists
```
- Shows all your playlists with IDs
- Quick access to manage playlists

**Delete Playlist**
```
/deleteplaylist [playlist_id]
```
- Permanently removes playlist
- All songs in playlist are removed

### Adding & Removing Songs

**Add to Playlist**
```
/addtoplaylist [playlist_id] [song_url]
```
- Supports: YouTube links, Spotify links, direct URLs
- Song metadata automatically saved
- Duplicate detection

**Remove from Playlist**
```
/removefromplaylist [playlist_id] [song_url]
```
- Remove specific song from playlist
- Preserves other playlist items

**View Playlist Info**
```
/playlistinfo [playlist_id]
```
- Shows all songs in playlist
- Display duration, artist info
- Total playlist duration

### Supported Playlist Formats

| Format | Support | Example |
|--------|---------|---------|
| YouTube Playlists | ✅ | `/play https://youtube.com/playlist?list=...` |
| Spotify Playlists | ✅ | `/play https://open.spotify.com/playlist/...` |
| User Playlists | ✅ | `/myplaylists` → Select → Play |

---

## 🌍 Supported Music Platforms

### Search & Download Support

The bot can search and play from multiple music platforms:

| Platform | Search | Direct Link | Metadata | Features |
|----------|--------|------------|----------|----------|
| **YouTube** | ✅ | ✅ | ✅ | Channels, playlists, clips |
| **Spotify** | ✅ | ✅ | ✅ | Artists, playlists, albums |
| **Apple Music** | ✅ | ✅ | ✅ | Albums, playlists |
| **SoundCloud** | ✅ | ✅ | ✅ | Tracks, sets |
| **Deezer** | ✅ | ✅ | ✅ | Albums, playlists |
| **JioSaavn** | ✅ | ✅ | ✅ | Indian music library |
| **Gaana** | ✅ | ✅ | ✅ | Indian songs, albums |
| **Tidal** | ✅ | ✅ | ✅ | High-quality audio |
| **Twitch** | ❌ | ✅ | ✅ | Clips, VODs |
| **Kick** | ❌ | ✅ | ✅ | Clips, streams |
| **Direct Links** | ❌ | ✅ | ✅ | Any HTTP audio file |

### URL Format Examples

```
YouTube:       https://youtube.com/watch?v=... or https://youtu.be/...
Spotify:       https://open.spotify.com/track/...
Apple Music:   https://music.apple.com/...
SoundCloud:    https://soundcloud.com/...
YouTube Music: https://music.youtube.com/...
Direct Link:   https://example.com/song.mp3
```

---

## ⭐ Feature Highlights

### Core Features

#### 🎵 **Multi-Platform Music Search**
- Search across 10+ music platforms
- Intelligent service routing
- Automatic format detection
- Fallback to alternative services

#### 🔊 **Advanced Playback Control**
- Adjustable playback speed (0.5x - 4.0x)
- Seek within tracks
- Loop individual songs
- Pause/resume functionality
- Mute/unmute controls

#### 📊 **Queue Management**
- View full playback queue
- Remove individual tracks
- Reorder tracks
- Persistent queue storage

#### 🎼 **Playlist System**
- Create unlimited playlists
- Add/remove songs
- Share playlists
- Persistent storage
- Quick access via buttons

#### 👥 **User Authorization**
- Admin-only mode
- Authorized user lists
- Granular permission control
- Blacklist/whitelist support

#### 🏠 **Group Settings**
- Per-group configurations
- Play mode restrictions
- Admin mode toggle
- Auto-delete command messages
- Customizable language (coming soon)

#### 🎤 **Voice Chat Integration**
- Join Telegram group voice chats
- Stream audio directly to group
- Automatic participant tracking
- Multi-chat support
- Low-latency streaming

#### 📈 **Performance Monitoring**
- System statistics tracking
- CPU/Memory usage monitoring
- Connected voice chats list
- Command execution logging

#### 🔐 **Security Features**
- User authentication
- Admin verification
- Blacklist management
- Secure session handling
- Permission hierarchy

### Advanced Features

#### 🚀 **Scalability**
- Support multiple sessions (accounts)
- Distribute load across assistants
- Session auto-rotation
- Connection pooling

#### 💾 **Caching System**
- In-memory cache for users
- Chat configuration cache
- Metadata caching
- Intelligent cache invalidation

#### 🔄 **Auto-Features**
- Auto-leave on errors
- Auto-cache reload
- Session auto-validation
- Connection auto-recovery

#### 📝 **Logging & Monitoring**
- Structured JSON logging
- Multiple log levels
- Per-module logging
- Runtime profiling (pprof)

#### 🌐 **Developer Tools**
- Shell command execution
- Broadcast messaging
- System statistics
- Active connection monitoring
- Assistant management

---

## 🎮 Interaction Examples

### Example 1: Playing a Song

```
User: /play Bohemian Rhapsody
Bot:  🎵 Searching for "Bohemian Rhapsody"...
      Searching on YouTube...
      ✅ Found: Bohemian Rhapsody - Queen
      ▶️ Now playing: Bohemian Rhapsody (5:55)
      👤 Added by: @username
```

### Example 2: Creating a Playlist

```
User: /createplaylist Rock Classics
Bot:  ✅ Playlist created: Rock Classics (ID: 42)

User: /addtoplaylist 42 https://youtube.com/watch?v=fJ9rUzIMt7o
Bot:  ✅ Added to playlist: Bohemian Rhapsody - Queen

User: /playlistinfo 42
Bot:  📊 Rock Classics Playlist
      Total Songs: 1
      Total Duration: 5:55
      Songs:
      1. Bohemian Rhapsody - Queen (5:55)
```

### Example 3: Admin Settings

```
User: /settings
Bot:  ⚙️ Group Settings
      [Play Mode: Everyone] [Admin Mode: Off]
      [Auto-Delete: Off] [Language: English]
      
      ← Back  Play Mode→
      
      (Admin taps "Play Mode→")
      
Bot:  Play Mode Settings
      Current: Everyone can play
      
      ✅ Everyone can play
      👤 Only admins can play
```

### Example 4: Speed Adjustment

```
User: /speed 1.5
Bot:  ⚙️ Playback speed changed to 1.5x
      Current track: "Song Name" (playing faster)
```

---

## 🔌 Keyboard Controls

The bot uses inline keyboards for easy interaction:

### Help Menu
```
[User] [Admin] [Owner] [Dev]
[Playlist] [Back]
```

### Settings Menu
```
[Play Mode] [Admin Mode] [Command Delete] [Language]
[Back]
```

### Playback Controls (Buttons)
```
⏮️ ⏸ ▶️ ⏭️
🔀 🔁 🔂 (Shuffle, Loop, Repeat)
🔇 🔊 (Volume)
```

---

## 📊 Statistics & Monitoring

The bot tracks various metrics:

- **Active Voice Chats:** Number of currently active voice connections
- **Total Users:** Registered users in database
- **Total Chats:** Configured groups/channels
- **Uptime:** Bot running duration
- **Memory Usage:** RAM consumption
- **CPU Usage:** Processor utilization
- **Session Status:** Active session validity
- **Database Status:** Connection health

View via: `/stats` (developer command)

---

## 🎯 Permission Hierarchy

```
Owner (Full access to all commands)
  ↓
Developers (dev commands + admin commands)
  ↓
Group Admins (admin commands in their group)
  ↓
Authorized Users (user commands)
  ↓
Everyone (basic user commands, configurable)
```

---

## 🔍 Command Aliases

Common aliases for convenience:

| Command | Alias |
|---------|-------|
| `/shell` | `/sh` |
| `/broadcast` | `/gcast` |
| `/active_vc` | `/av` |

---

## 📌 Important Notes

- **URL Support:** Most commands accept song names (searches) or direct URLs
- **Duration Limits:** Owners can set maximum song duration
- **Permissions:** Admin settings override user settings
- **Database:** All user data persists across bot restarts
- **Sessions:** Multiple accounts can be configured for load distribution
- **Compatibility:** Works with Telegram's latest API

---

## 🆘 FAQ

**Q: Can the bot play videos in voice chat?**
A: Yes, videos are converted to audio and streamed.

**Q: What's the maximum queue size?**
A: Unlimited (depends on available memory).

**Q: Can I create private playlists?**
A: Yes, playlists are automatically private to your account.

**Q: Does the bot support offline songs?**
A: No, it requires internet to stream or download songs.

**Q: Can I restrict commands by permission level?**
A: Yes, use `/settings` for granular control.

---

**Last Updated:** 2026-05-24
