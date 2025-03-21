package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

type PostgresDB struct {
	db *sql.DB
}

type PostgresConfig struct {
	DSN string
}

func NewPostgresDB(cfg PostgresConfig) (*PostgresDB, error) {

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

func NewPostgresDBFromEnv() (*PostgresDB, error) {
	dsn := os.Getenv("POSTGRES_DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("missing POSTGRES_DB_DSN environment variable")
	}
	cfg := PostgresConfig{
		DSN: dsn,
	}
	return NewPostgresDB(cfg)
}

func (p *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

func (p *PostgresDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return rows, nil
}

func (p *PostgresDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := p.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return result, nil
}

func (p *PostgresDB) Close() error {
	err := p.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

func (p *PostgresDB) Ping(ctx context.Context) error {
	err := p.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := p.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (p *PostgresDB) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := p.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	return stmt, nil
}

func (p *PostgresDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	stmt, err := p.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare context statement: %w", err)
	}
	return stmt, nil
}

func (p *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute context query: %w", err)
	}
	return result, nil
}

func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute context query: %w", err)
	}
	return rows, nil
}

func (p *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row := p.db.QueryRowContext(ctx, query, args...)
	return row
}

func (p *PostgresDB) MustExec(query string, args ...interface{}) {
	_, err := p.Exec(query, args...)
	if err != nil {
		log.Fatalf("MustExec failed: %v", err)
	}
}
