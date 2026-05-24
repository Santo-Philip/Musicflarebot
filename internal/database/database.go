package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(databaseUrl string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	DB = pool
	slog.Info("connected to postgres")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func CreateTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS bot_settings (id SERIAL PRIMARY KEY, key TEXT UNIQUE NOT NULL, value TEXT NOT NULL DEFAULT '', updated_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, user_id BIGINT UNIQUE NOT NULL, username TEXT DEFAULT '', first_name TEXT DEFAULT '', last_name TEXT DEFAULT '', language TEXT DEFAULT 'en', is_bot_allowed BOOLEAN DEFAULT true, ban_permanent BOOLEAN DEFAULT false, banned_until TIMESTAMP, created_at TIMESTAMP DEFAULT NOW(), updated_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS chats (id SERIAL PRIMARY KEY, chat_id BIGINT UNIQUE NOT NULL, title TEXT DEFAULT '', username TEXT DEFAULT '', chat_type TEXT DEFAULT '', is_forum BOOLEAN DEFAULT false, is_bot_allowed BOOLEAN DEFAULT true, created_at TIMESTAMP DEFAULT NOW(), updated_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS user_stats (id SERIAL PRIMARY KEY, user_id BIGINT NOT NULL, date DATE NOT NULL DEFAULT CURRENT_DATE, total_requests INTEGER DEFAULT 0, total_music_requests INTEGER DEFAULT 0, total_video_requests INTEGER DEFAULT 0, total_successful_requests INTEGER DEFAULT 0, total_failed_requests INTEGER DEFAULT 0, total_playlist_imports INTEGER DEFAULT 0, total_playlist_exports INTEGER DEFAULT 0, UNIQUE(user_id, date))`,
		`CREATE TABLE IF NOT EXISTS playlists (id SERIAL PRIMARY KEY, user_id BIGINT NOT NULL, playlist_name TEXT NOT NULL, song_track JSONB DEFAULT '[]', created_at TIMESTAMP DEFAULT NOW(), updated_at TIMESTAMP DEFAULT NOW(), UNIQUE(user_id, playlist_name))`,
		`CREATE TABLE IF NOT EXISTS assistants (id SERIAL PRIMARY KEY, user_id BIGINT NOT NULL, name TEXT DEFAULT 'Assistant', session_string TEXT NOT NULL, api_id INTEGER DEFAULT 0, api_hash TEXT DEFAULT '', is_logged_in BOOLEAN DEFAULT false, created_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS blacklist_chats (id SERIAL PRIMARY KEY, chat_id BIGINT UNIQUE NOT NULL, reason TEXT DEFAULT '', created_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS blacklist_users (id SERIAL PRIMARY KEY, user_id BIGINT UNIQUE NOT NULL, reason TEXT DEFAULT '', created_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS auth_users (id SERIAL PRIMARY KEY, user_id BIGINT UNIQUE NOT NULL, permissions TEXT DEFAULT '{}', created_at TIMESTAMP DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS lang (id SERIAL PRIMARY KEY, code TEXT UNIQUE NOT NULL, data JSONB DEFAULT '{}')`,
		`CREATE TABLE IF NOT EXISTS logger_status (id SERIAL PRIMARY KEY, group_id BIGINT UNIQUE NOT NULL, user_id BIGINT NOT NULL, status TEXT DEFAULT 'idle' CHECK(status IN ('idle','logging','paused')))`,
	}
	for _, q := range tables {
		if _, err := DB.Exec(context.Background(), q); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}
	slog.Info("database tables ready")
	return nil
}

func GetSetting(key string) (string, error) {
	var v string
	err := DB.QueryRow(context.Background(), "SELECT value FROM bot_settings WHERE key = $1", key).Scan(&v)
	return v, err
}

func SetSetting(key, value string) error {
	_, err := DB.Exec(context.Background(),
		`INSERT INTO bot_settings (key, value, updated_at) VALUES ($1,$2,NOW()) ON CONFLICT (key) DO UPDATE SET value=$2,updated_at=NOW()`,
		key, value)
	return err
}

func GetAllSettings() (map[string]string, error) {
	rows, err := DB.Query(context.Background(), "SELECT key, value FROM bot_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}

func IsUserBanned(userId int64) bool {
	var perm bool
	var until time.Time
	err := DB.QueryRow(context.Background(),
		"SELECT ban_permanent, banned_until FROM users WHERE user_id = $1", userId,
	).Scan(&perm, &until)
	if err != nil {
		return false
	}
	return perm || (!until.IsZero() && time.Now().Before(until))
}
