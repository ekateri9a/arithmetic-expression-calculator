package models

import (
	"arithmetic-expression-calculator/internal/logger"
	"context"
	"database/sql"
	"strconv"
)

type Expression struct {
	ID         int64   `json:"id"`
	Expression string  `json:"expression"`
	Status     string  `json:"status"`
	Result     float64 `json:"result"`
	UserID     int64   `json:"-"`
}

func (e Expression) Print() string {
	id := strconv.FormatInt(e.ID, 10)
	userID := strconv.FormatInt(e.UserID, 10)
	result := strconv.FormatFloat(e.Result, 'g', -1, 64)
	return "ID: " + id + " Expression:" + e.Expression + " Status:" + e.Status + " Result:" + result + " UserID:" + userID
}

func InsertExpression(ctx context.Context, db *sql.DB, expression *Expression) (int64, error) {
	var q = `
	INSERT INTO expressions (expression, status, user_id, result) values ($1, $2, $3, $4)
	`
	result, err := db.ExecContext(ctx, q, expression.Expression, "calculate", expression.UserID, 0)
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	return id, nil
}

func SelectExpressionsByUserID(ctx context.Context, db *sql.DB, userID int64) ([]Expression, error) {
	var expressions []Expression
	var q = "SELECT * FROM expressions where user_id = $1"

	rows, err := db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := Expression{}
		err := rows.Scan(&e.ID, &e.Expression, &e.Status, &e.Result, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

func SelectExpressionByUserID(ctx context.Context, db *sql.DB, userID int64, id int64) (Expression, error) {
	e := Expression{}
	var q = "SELECT * FROM expressions where id = $1 and user_id = $2"
	err := db.QueryRowContext(ctx, q, id, userID).Scan(&e.ID, &e.Expression, &e.Status, &e.Result, &e.UserID)
	if err != nil {
		return e, err
	}

	return e, nil
}

func UpdateExpressionFinish(ctx context.Context, db *sql.DB, id int64, result float64) error {
	status := "finished"
	var q = "UPDATE expressions SET status = $1, result = $2 WHERE id = $3"
	_, err := db.ExecContext(ctx, q, status, result, id)
	if err != nil {
		return err
	}

	return nil
}
