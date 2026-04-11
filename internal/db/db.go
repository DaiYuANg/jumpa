package db

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/DaiYuANg/arcgo/dbx"
	mysql "github.com/DaiYuANg/arcgo/dbx/dialect/mysql"
	postgres "github.com/DaiYuANg/arcgo/dbx/dialect/postgres"
	sqlitedialect "github.com/DaiYuANg/arcgo/dbx/dialect/sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func Open(driver, dsn string, opts ...dbx.Option) (*dbx.DB, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "mysql", "mariadb":
		return OpenMySQL(dsn, opts...)
	case "postgres", "postgresql":
		return OpenPostgres(dsn, opts...)
	case "", "sqlite":
		return OpenSQLite(dsn, opts...)
	default:
		return nil, fmt.Errorf("unsupported db driver %q, supported: sqlite, mysql, mariadb, postgres", driver)
	}
}

func OpenSQLite(dsn string, opts ...dbx.Option) (*dbx.DB, error) {
	if dsn == "" {
		dsn = "file:backend?mode=memory&cache=shared"
	}
	db, err := dbx.Open(
		dbx.WithDriver("sqlite"),
		dbx.WithDSN(dsn),
		dbx.WithDialect(sqlitedialect.New()),
		dbx.ApplyOptions(opts...),
	)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func OpenMySQL(dsn string, opts ...dbx.Option) (*dbx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is required")
	}
	db, err := dbx.Open(
		dbx.WithDriver("mysql"),
		dbx.WithDSN(dsn),
		dbx.WithDialect(mysql.New()),
		dbx.ApplyOptions(opts...),
	)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	return db, nil
}

func OpenPostgres(dsn string, opts ...dbx.Option) (*dbx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	db, err := dbx.Open(
		dbx.WithDriver("pgx"),
		dbx.WithDSN(dsn),
		dbx.WithDialect(postgres.New()),
		dbx.ApplyOptions(opts...),
	)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return db, nil
}

func DefaultOpts(logger *slog.Logger) []dbx.Option {
	if logger == nil {
		return nil
	}
	return []dbx.Option{
		dbx.WithLogger(logger),
		dbx.WithDebug(false),
	}
}
