package db

import (
	"musicflarebot/src/utils"
	"log/slog"
)

type Chats struct {
	ID        int64
	PlayType  int
	AdminPlay bool
	AdminMode string
	CmdDelete bool
}

func (db *Database) getChat(chatID int64) (*Chats, error) {
	key := toKey(chatID)
	if cached, ok := db.chatCache.Get(key); ok {
		return cached, nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var chat Chats
	err := db.pool.QueryRow(ctx,
		`SELECT id, play_type, admin_play, admin_mode, cmd_delete FROM chats WHERE id = $1`, chatID,
	).Scan(&chat.ID, &chat.PlayType, &chat.AdminPlay, &chat.AdminMode, &chat.CmdDelete)

	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		slog.Info("[DB] Error getting chat", "error", err)
		return nil, err
	}

	db.chatCache.Set(key, &chat)
	return &chat, nil
}

func (db *Database) AddChat(chatID int64) error {
	chat, _ := db.getChat(chatID)
	if chat != nil {
		return nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, chatID,
	)
	if err == nil {
		slog.Info("[DB] A new chat has been added", "id", chatID)
	}
	return err
}

func (db *Database) GetPlayType(chatID int64) int {
	chat, _ := db.getChat(chatID)
	if chat == nil {
		return 0
	}
	return chat.PlayType
}

func (db *Database) SetPlayType(chatID int64, playType int) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, play_type) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET play_type = $2`,
		chatID, playType,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetPlayMode(chatID int64) bool {
	chat, _ := db.getChat(chatID)
	if chat == nil {
		return false
	}
	return chat.AdminPlay
}

func (db *Database) SetPlayMode(chatID int64, adminPlay bool) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, admin_play) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET admin_play = $2`,
		chatID, adminPlay,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetAdminMode(chatID int64) string {
	chat, _ := db.getChat(chatID)
	if chat == nil || chat.AdminMode == "" {
		return utils.Everyone
	}
	return chat.AdminMode
}

func (db *Database) SetAdminMode(chatID int64, adminMode string) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, admin_mode) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET admin_mode = $2`,
		chatID, adminMode,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetCmdDelete(chatID int64) bool {
	chat, _ := db.getChat(chatID)
	if chat == nil {
		return false
	}
	return chat.CmdDelete
}

func (db *Database) SetCmdDelete(chatID int64, cmdDelete bool) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, cmd_delete) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET cmd_delete = $2`,
		chatID, cmdDelete,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetAllChats() ([]int64, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT id FROM chats`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		chats = append(chats, id)
	}
	return chats, rows.Err()
}
