package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func toKey(id int64) string {
	return fmt.Sprintf("%d", id)
}

func contains(list []int64, id int64) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
