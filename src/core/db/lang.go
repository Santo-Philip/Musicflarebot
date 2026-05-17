package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (db *Database) GetLanguage(chatID int64) (string, error) {
	key := toKey(chatID)
	if cached, ok := db.langCache.Get(key); ok {
		return cached, nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var lang string
	err := db.pool.QueryRow(ctx,
		`SELECT lang FROM chat_languages WHERE chat_id = $1`, chatID,
	).Scan(&lang)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "en", nil
		}
		return "", err
	}

	db.langCache.Set(key, lang)
	return lang, nil
}

func (db *Database) SetLanguage(ctx context.Context, chatID int64, langCode string) error {
	_, err := db.pool.Exec(ctx,
		`INSERT INTO chat_languages (chat_id, lang) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET lang = $2`,
		chatID, langCode,
	)
	if err == nil {
		db.langCache.Set(toKey(chatID), langCode)
	}
	return err
}
