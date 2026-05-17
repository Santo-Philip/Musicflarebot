package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Users struct {
	ID int64
}

func (db *Database) AddUser(userID int64) error {
	key := toKey(userID)
	if _, ok := db.userCache.Get(key); ok {
		return nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, userID,
	)
	if err != nil {
		return err
	}

	db.userCache.Set(key, &Users{ID: userID})
	return nil
}

func (db *Database) RemoveUser(userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return err
	}

	db.userCache.Delete(toKey(userID))
	return nil
}

func (db *Database) IsUserExist(userID int64) (bool, error) {
	key := toKey(userID)
	if _, ok := db.userCache.Get(key); ok {
		return true, nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var id int64
	err := db.pool.QueryRow(ctx, `SELECT id FROM users WHERE id = $1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	db.userCache.Set(key, &Users{ID: id})
	return true, nil
}

func (db *Database) GetAllUsers() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT id FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		users = append(users, id)
	}
	return users, rows.Err()
}
