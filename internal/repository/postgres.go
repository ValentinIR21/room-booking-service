package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

// Подключение к БД
func ConnectionDB(ctx context.Context, connString string) (*DB, error) {

	//Конфиг пула
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("err: %w", err)
	}

	// Создание пула
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Не удалось создать pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("БД не отвечает: %w", err)
	}

	return &DB{
		Pool: pool,
	}, nil

}

// Закрытие соединения с БД
func (db *DB) CloseDB() {
	db.Pool.Close()
}
