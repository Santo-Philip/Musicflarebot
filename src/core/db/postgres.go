package db

import (
	"musicflarebot/config"
	"musicflarebot/src/core/cache"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool

	chatCache      *cache.Cache[*Chats]
	userCache      *cache.Cache[*Users]
	assistantCache *cache.Cache[int]
	authCache      *cache.Cache[[]int64]
	langCache      *cache.Cache[string]
	loggerCache    *cache.Cache[bool]
	blChatsCache   *cache.Cache[[]int64]
	blUsersCache   *cache.Cache[[]int64]
}

var Instance *Database

func InitDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.Conf.DatabaseUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Instance = &Database{
		pool: pool,

		chatCache:      cache.NewCache[*Chats](20 * time.Minute),
		userCache:      cache.NewCache[*Users](20 * time.Minute),
		assistantCache: cache.NewCache[int](20 * time.Minute),
		authCache:      cache.NewCache[[]int64](20 * time.Minute),
		langCache:      cache.NewCache[string](20 * time.Minute),
		loggerCache:    cache.NewCache[bool](20 * time.Minute),
		blChatsCache:   cache.NewCache[[]int64](20 * time.Minute),
		blUsersCache:   cache.NewCache[[]int64](20 * time.Minute),
	}

	if err := Instance.migrate(); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	slog.Info("[DB] PostgreSQL connection established successfully.")
	return nil
}

func (db *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chats (
			id BIGINT PRIMARY KEY,
			play_type INT DEFAULT 0,
			admin_play BOOLEAN DEFAULT FALSE,
			admin_mode TEXT DEFAULT 'everyone',
			cmd_delete BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY
		)`,
		`CREATE TABLE IF NOT EXISTS assistant_assignments (
			chat_id BIGINT PRIMARY KEY,
			num INT DEFAULT -1
		)`,
		`CREATE TABLE IF NOT EXISTS auth_users (
			chat_id BIGINT,
			user_id BIGINT,
			PRIMARY KEY (chat_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS blacklist (
			type TEXT,
			id BIGINT,
			PRIMARY KEY (type, id)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value JSONB DEFAULT '{}'::jsonb
		)`,
		`CREATE TABLE IF NOT EXISTS playlists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			user_id BIGINT NOT NULL,
			songs JSONB DEFAULT '[]'::jsonb
		)`,
		`CREATE TABLE IF NOT EXISTS chat_languages (
			chat_id BIGINT PRIMARY KEY,
			lang TEXT DEFAULT 'en'
		)`,
		`CREATE TABLE IF NOT EXISTS bot_settings (
			s_key TEXT PRIMARY KEY,
			s_value TEXT NOT NULL DEFAULT ''
		)`,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, q := range queries {
		if _, err := tx.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration failed on query: %s: %w", q, err)
		}
	}

	return tx.Commit(ctx)
}

func (db *Database) Close() {
	db.pool.Close()
	slog.Info("[DB] Database connection closed.")
}

func (db *Database) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

var ErrNotFound = errors.New("not found")
