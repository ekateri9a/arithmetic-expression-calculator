package models

import (
	"context"
	"database/sql"
	"strconv"
)

type User struct {
	ID       int64
	Login    string
	Password string
}

func (u User) Print() string {
	id := strconv.FormatInt(u.ID, 10)
	return "ID: " + id + " Name: " + u.Login
}

func InsertUser(ctx context.Context, db *sql.DB, user *User) (int64, error) {
	var q = `
	INSERT INTO users (login, password) values ($1, $2)
	`
	result, err := db.ExecContext(ctx, q, user.Login, user.Password)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func SelectUserByLogin(ctx context.Context, db *sql.DB, login string) (User, error) {
	u := User{}
	var q = "SELECT * FROM users WHERE login = $1"
	err := db.QueryRowContext(ctx, q, login).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		return u, err
	}

	return u, nil
}
