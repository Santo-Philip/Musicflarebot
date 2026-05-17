package db

import (
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func (db *Database) GetAssistant(chatID int64) (int, error) {
	key := toKey(chatID)
	if cached, ok := db.assistantCache.Get(key); ok {
		return cached, nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var num int
	err := db.pool.QueryRow(ctx,
		`SELECT num FROM assistant_assignments WHERE chat_id = $1`, chatID,
	).Scan(&num)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, nil
		}
		return -1, err
	}

	db.assistantCache.Set(key, num)
	return num, nil
}

func (db *Database) SetAssistant(chatID int64, num int) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO assistant_assignments (chat_id, num) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET num = $2`,
		chatID, num,
	)
	if err == nil {
		db.assistantCache.Set(toKey(chatID), num)
	}
	return err
}

func (db *Database) RemoveAssistant(chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM assistant_assignments WHERE chat_id = $1`, chatID,
	)
	if err == nil {
		db.assistantCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) AssignAssistant(chatID int64, proposedAssistant int) (int, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback(ctx)

	var existingNum int
	err = tx.QueryRow(ctx,
		`SELECT num FROM assistant_assignments WHERE chat_id = $1 FOR UPDATE`, chatID,
	).Scan(&existingNum)

	if err == nil && existingNum != -1 {
		tx.Commit(ctx)
		return existingNum, nil
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO assistant_assignments (chat_id, num) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET num = $2`,
		chatID, proposedAssistant,
	)
	if err != nil {
		return -1, err
	}

	if err := tx.Commit(ctx); err != nil {
		return -1, err
	}

	db.assistantCache.Set(toKey(chatID), proposedAssistant)
	return proposedAssistant, nil
}

func (db *Database) ClearAllAssistants() (int64, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	result, err := db.pool.Exec(ctx, `DELETE FROM assistant_assignments`)
	if err != nil {
		slog.Info("[DB] Error clearing assistants", "error", err)
		return 0, err
	}

	db.assistantCache.Clear()
	return result.RowsAffected(), nil
}
