package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier — общий интерфейс для pgxpool.Pool и pgx.Tx.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type txKey struct{}

// WithTxContext возвращает контекст с вложенной транзакцией.
func WithTxContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

type DB struct {
	Pool *pgxpool.Pool
}

// Подключение к БД
func ConnectionDB(ctx context.Context, connString string) (*DB, error) {

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &DB{
		Pool: pool,
	}, nil

}

// WithTx выполняет fn внутри транзакции. Коммитит при успехе, откатывает при ошибке.
// Транзакция доступна вложенным вызовам через db.Conn(ctx).
func (db *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := WithTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Conn возвращает транзакцию из контекста или пул.
func (db *DB) Conn(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.Pool
}

// Закрытие соединения с БД
func (db *DB) CloseDB() {
	db.Pool.Close()
}
