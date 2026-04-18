package database

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
    "github.com/Kerefall/mobile-service-engineer/internal/config"
)

type PostgresDB struct {
    Pool *pgxpool.Pool
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
    // Строка подключения
    connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBHost,
        cfg.DBPort,
        cfg.DBName,
    )

    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    config.MaxConns = 10
    config.MinConns = 2

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        return nil, fmt.Errorf("failed to create pool: %w", err)
    }

    if err := pool.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    logrus.Info("Connected to PostgreSQL")
    return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Close() {
    if db.Pool != nil {
        db.Pool.Close()
        logrus.Info("PostgreSQL connection closed")
    }
}