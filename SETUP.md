# MusicFlareBot - Setup Guide

An intelligent Telegram music player bot built with Go, featuring voice chat integration, multi-platform music support, and advanced playback controls.

**Developer:** [@NexFang](https://t.me/nexfang)

---

## 📋 Table of Contents

1. [System Requirements](#system-requirements)
2. [Project Structure](#project-structure)
3. [Environment Configuration](#environment-configuration)
4. [Database Setup](#database-setup)
5. [Installation & Deployment](#installation--deployment)
6. [Development Setup](#development-setup)
7. [Architecture Overview](#architecture-overview)
8. [Configuration Details](#configuration-details)

---

## 🖥️ System Requirements

### Minimum Requirements
- **Go:** 1.26 or later
- **PostgreSQL:** 12 or later
- **FFmpeg:** For audio/video processing
- **RAM:** 512MB minimum (1GB+ recommended)
- **Disk:** 2GB minimum (for cache and downloads)

### External Dependencies
- **Telegram API:** BotAPI Token and App credentials
- **Music Service API:** API Key for music metadata service
- **ntgcalls:** Voice chat integration library (automatically built)

---

## 📁 Project Structure

```
musicflarebot/
├── config/                 # Configuration management
│   ├── config.go          # Config loader
│   ├── load.go            # Environment variable loading
│   └── types.go           # Config type definitions
│
├── src/
│   ├── init.go            # Bot initialization
│   ├── core/              # Core functionality
│   │   ├── buttons.go     # Keyboard markup generation
│   │   ├── cache/         # Memory caching layer
│   │   │   ├── cache.go
│   │   │   ├── admin_cache.go
│   │   │   └── chat_cache.go
│   │   ├── db/            # Database operations
│   │   │   ├── postgres.go
│   │   │   ├── users.go
│   │   │   ├── chats.go
│   │   │   ├── playlist.go
│   │   │   ├── assistant.go
│   │   │   ├── auth.go
│   │   │   ├── blacklist.go
│   │   │   ├── settings.go
│   │   │   └── ...
│   │   └── dl/            # Download & search functionality
│   │       ├── downloader.go
│   │       ├── api.go
│   │       ├── youtube.go
│   │       ├── spotify_dl.go
│   │       ├── direct_link.go
│   │       ├── youtube_search.go
│   │       ├── http_client.go
│   │       └── ...
│   │
│   ├── handlers/          # Command handlers (35+ handlers)
│   │   ├── play.go        # /play command
│   │   ├── skip.go        # /skip command
│   │   ├── pause.go       # /pause command
│   │   ├── queue.go       # /queue display
│   │   ├── playlist.go    # Playlist management
│   │   ├── loop.go        # Loop functionality
│   │   ├── seek.go        # Seek in track
│   │   ├── speed.go       # Playback speed
│   │   ├── mute.go        # Mute/unmute
│   │   ├── settings.go    # Chat settings
│   │   ├── auth.go        # User authentication
│   │   ├── admins.go      # Admin commands
│   │   ├── owner_settings.go  # Owner settings
│   │   ├── devs.go        # Developer commands
│   │   ├── stats.go       # System statistics
│   │   ├── shell.go       # Shell execution
│   │   ├── broadcast.go   # Broadcast messages
│   │   └── ...
│   │
│   ├── utils/             # Utility functions
│   │   ├── models.go      # Data structures
│   │   ├── telegram.go    # Telegram utilities
│   │   ├── regex.go       # Regex patterns
│   │   └── durations.go   # Duration parsing
│   │
│   └── vc/                # Voice chat module
│       ├── calls.go       # Voice call management
│       ├── start.go       # Call startup
│       ├── helpers.go     # Helper functions
│       ├── types.go       # Type definitions
│       ├── ntgcalls/      # ntgcalls bindings
│       │   ├── client.go
│       │   ├── protocol.go
│       │   ├── types.go
│       │   └── ...
│       └── ubot/          # Userbot utilities
│           ├── calls.go
│           ├── connect_call.go
│           ├── context.go
│           └── ...
│
├── main.go                # Entry point
├── setup_ntgcalls.go      # ntgcalls setup
├── go.mod                 # Go module definition
├── Dockerfile             # Docker build configuration
├── docker-compose.yml     # Docker Compose setup
├── sample.env             # Environment template
└── LICENSE                # License information
```

---

## ⚙️ Environment Configuration

### Required Environment Variables

Create a `.env` file in the project root with the following variables:

```bash
# Telegram API Credentials (Required)
API_ID=<your_api_id>                    # Get from https://my.telegram.org/apps
API_HASH=<your_api_hash>                # Get from https://my.telegram.org/apps
TOKEN=<your_bot_token>                  # Get from @BotFather
OWNER_ID=<owner_user_id>                # Your Telegram user ID

# Database Configuration (Required)
DATABASE_URL=postgresql://user:password@localhost:5432/musicflarebot

# Music Service API (Required)
API_URL=https://tgmusic.fallenapi.fun   # Base URL for music metadata API
API_KEY=<your_api_key>                  # API key for music service

# Session Management (Optional)
SESSION_TYPE=gogram                     # Session type (gogram or telethon)
SESSION=<session_string>                # Session string (alternative to database)
SESSION1=<session_string>               # Additional session strings (SESSION2, SESSION3, etc.)

# Optional Configuration
PORT=6060                               # Profiling server port
LOGGER_ID=<logger_channel_id>          # Channel for bot logs
SUPPORT_GROUP=<support_group_username>  # Support group username
SUPPORT_CHANNEL=<support_channel_id>   # Support channel identifier
```

### Configuration via Database

Owner settings can also be configured via commands:
- `/set_logger_id [id]`
- `/set_support_group [username]`
- `/set_support_channel [username]`
- `/set_song_duration [seconds]`
- `/set_default_service [service]`

---

## 🗄️ Database Setup

### PostgreSQL Schema

The bot automatically creates the following tables on initialization:

#### Core Tables
- **users** - User accounts and preferences
- **chats** - Group/channel configurations
- **sessions** - Session management
- **playlists** - User-created playlists
- **settings** - Chat and user settings
- **auth** - Authorization records
- **blacklist** - Blocked users/chats
- **assistants** - Assistant account assignments
- **cache** - Cached data

### Database Connection

```bash
# Local PostgreSQL
DATABASE_URL=postgresql://postgres:password@localhost:5432/musicflarebot

# Docker (via docker-compose)
DATABASE_URL=postgresql://musicflare:musicflare@postgres:5432/musicflarebot

# Remote PostgreSQL
DATABASE_URL=postgresql://user:password@host:5432/musicflarebot
```

### Database Initialization

```bash
# Using go-migrate (recommended)
migrate -path migrations -database $DATABASE_URL up

# Automatic initialization (via db package)
# The bot will automatically create tables on first run
```

---

## 🚀 Installation & Deployment

### Option 1: Local Development

#### Prerequisites
```bash
# Install Go 1.26+
# Install PostgreSQL 12+
# Install FFmpeg
```

#### Setup Steps

1. **Clone Repository**
   ```bash
   git clone <repository-url>
   cd musicflarebot
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   go run setup_ntgcalls.go
   ```

3. **Configure Environment**
   ```bash
   cp sample.env .env
   # Edit .env with your credentials
   ```

4. **Start PostgreSQL**
   ```bash
   # Ensure PostgreSQL is running on localhost:5432
   psql -U postgres -c "CREATE DATABASE musicflarebot;"
   ```

5. **Run the Bot**
   ```bash
   go run main.go
   ```

### Option 2: Docker Deployment (Recommended)

#### Prerequisites
```bash
# Install Docker 20.10+
# Install Docker Compose 2.0+
```

#### Setup Steps

1. **Configure Environment**
   ```bash
   cp sample.env .env
   # Edit .env with your credentials
   ```

2. **Build & Run**
   ```bash
   # Build and start all services
   docker-compose up --build

   # Run in background
   docker-compose up -d --build

   # View logs
   docker-compose logs -f musicflarebot
   ```

3. **Verify Installation**
   ```bash
   docker-compose ps
   # Both musicflarebot and postgres should be running
   ```

4. **Stop Services**
   ```bash
   docker-compose down

   # Stop and remove volumes
   docker-compose down -v
   ```

#### Docker Services

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| musicflarebot | golang:1.26-bookworm | 6060 (pprof) | Bot application |
| postgres | postgres:17-alpine | 5432 | Database |

### Option 3: Production Deployment

#### Systemd Service

Create `/etc/systemd/system/musicflarebot.service`:

```ini
[Unit]
Description=MusicFlareBot
After=network.target postgresql.service

[Service]
Type=simple
User=musicflare
WorkingDirectory=/opt/musicflarebot
ExecStart=/usr/local/bin/musicflarebot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

#### Reverse Proxy (Nginx)

```nginx
server {
    listen 6060;
    server_name bot.example.com;

    location / {
        proxy_pass http://localhost:6060;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 🔧 Development Setup

### Local Development Environment

#### Using VS Code

1. **Install Extensions**
   - Go (golang.go)
   - Database Client (cweijan.vscode-database-client2)
   - SQLTools (mtxr.sqltools)

2. **Launch Configuration**
   
   Create `.vscode/launch.json`:
   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "MusicFlareBot",
         "type": "go",
         "request": "launch",
         "mode": "debug",
         "program": "${workspaceFolder}",
         "cwd": "${workspaceFolder}",
         "env": {
           "DATABASE_URL": "postgresql://postgres:password@localhost:5432/musicflarebot"
         }
       }
     ]
   }
   ```

#### Using Hot Reload

```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Create .air.toml in project root
air

# Bot will restart on file changes
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Test specific package
go test ./src/core/cache -v
```

### Building

```bash
# Development build
go build -o musicflarebot main.go

# Production build (optimized)
CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o musicflarebot main.go

# Cross-compile for different platforms
GOOS=darwin GOARCH=amd64 go build -o musicflarebot.macos main.go
GOOS=windows GOARCH=amd64 go build -o musicflarebot.exe main.go
```

---

## 🏗️ Architecture Overview

### Component Interaction

```
┌─────────────────────────────────────────┐
│         Telegram User/Group             │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│      Telegram Bot API (gogram)          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────┐
│              Handler Layer (35+ handlers)           │
│  Play, Skip, Pause, Queue, Playlist, Settings...   │
└──────────────┬──────────────────────────────────────┘
               │
      ┌────────┴────────┬──────────────┬──────────────┐
      │                 │              │              │
┌─────▼──────┐  ┌──────▼──────┐  ┌───▼──────┐  ┌───▼──────┐
│ Core Cache │  │  Database   │  │ Download │  │ Voice    │
│ (In-Memory)│  │ (PostgreSQL)│  │ & Search │  │ Chat     │
└────────────┘  └─────────────┘  └──────────┘  └─────────┘
      │                 │              │              │
┌─────▼──────┬─────────▼──────┬──────▼──────┬──────▼──────┐
│  User &    │  Playlist &    │ YouTube,   │ ntgcalls   │
│  Chat      │  Settings      │ Spotify,   │ (Voice)    │
│  Cache     │  Mgmt          │ APIs, DL   │ ubot       │
└────────────┴────────────────┴────────────┴────────────┘
```

### Data Flow

1. **Command Reception** → Handler determines action
2. **Permission Check** → Verify user/admin permissions
3. **Cache Check** → Look up cached data
4. **Database Query** → Fetch or store data
5. **API Call** → Download/search music metadata
6. **Voice Chat** → Stream audio to group
7. **Response** → Send status to user

### Module Responsibilities

| Module | Responsibility |
|--------|-----------------|
| **config** | Load and manage configuration from environment |
| **handlers** | Process user commands and callbacks |
| **core/cache** | In-memory caching of users, chats, and metadata |
| **core/db** | PostgreSQL operations and data persistence |
| **core/dl** | Music search and download operations |
| **src/vc** | Voice chat connection and audio streaming |
| **utils** | Utility functions and data models |

---

## 📝 Configuration Details

### Bot Configuration Structure

```go
type BotConfig struct {
    ApiId              int32     // Telegram API ID
    ApiHash            string    // Telegram API Hash
    Token              string    // Bot token from @BotFather
    SessionStrings     []string  // Multiple session strings
    SessionType        string    // "gogram" or "telethon"
    DatabaseUrl        string    // PostgreSQL connection string
    ApiUrl             string    // Music metadata API base URL
    ApiKey             string    // API authentication key
    OwnerId            int64     // Bot owner's user ID
    LoggerId           int64     // Logger channel ID
    Proxy              string    // Optional proxy configuration
    DefaultService     string    // Default music streaming service
    MaxFileSize        int64     // Maximum file size (bytes)
    SongDurationLimit  int64     // Max song duration (seconds)
    DownloadsDir       string    // Cache directory path
    SupportGroup       string    // Support group username
    SupportChannel     string    // Support channel ID
    DEVS               []int64   // Developer user IDs
    CookiesPath        []string  // Cookie file paths
    StartImg           string    // Startup image URL
    Port               string    // Profiling server port
    AutoLeave          bool      // Auto-leave on error flag
}
```

### Supported Music Platforms

| Platform | Support | Features |
|----------|---------|----------|
| YouTube | ✅ | Search, play, playlist |
| Spotify | ✅ | Track info, playlist |
| Apple Music | ✅ | Track metadata |
| SoundCloud | ✅ | Search, play |
| Deezer | ✅ | Track info |
| JioSaavn | ✅ | Indian music library |
| Gaana | ✅ | Indian songs |
| Tidal | ✅ | High-quality audio |
| Twitch | ✅ | Stream clips |
| Kick | ✅ | Stream clips |

---

## 📚 Additional Resources

### Official Documentation
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [gogram Documentation](https://github.com/amarnathcjd/gogram)
- [ntgcalls Documentation](https://github.com/pytgcalls/ntgcalls)

### Database
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [pgx Driver Documentation](https://github.com/jackc/pgx)

### Commands Reference
See [FEATURES.md](FEATURES.md) for complete command reference and feature documentation.

---

## 🆘 Troubleshooting

### Common Issues

**Issue:** Bot doesn't start
```
Solution: Check .env file, verify DATABASE_URL, ensure PostgreSQL is running
```

**Issue:** Voice chat not working
```
Solution: Verify ntgcalls setup, check session strings are valid
```

**Issue:** Database connection error
```
Solution: Check PostgreSQL is running, verify DATABASE_URL, check network connectivity
```

**Issue:** Songs not playing
```
Solution: Check API_KEY and API_URL, verify FFmpeg is installed, check file permissions
```

### Debugging

Enable debug logging:
```bash
# Via environment
export RUST_LOG=debug
go run main.go

# Via code modification in main.go
slog.SetDefault(logger)  // Verify log handler level
```

---

## 📄 License

This project is licensed under the same terms as the ntgcalls library (GNU Lesser General Public License v3.0).

---

**Last Updated:** 2026-05-24
**Version:** Current Development Build
