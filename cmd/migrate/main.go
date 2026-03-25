package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	config2 "github.com/DaiYuANg/arcgo-rbac-template/internal/config"
	db2 "github.com/DaiYuANg/arcgo-rbac-template/internal/db"
	"github.com/DaiYuANg/arcgo/configx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/dialect"
	"github.com/DaiYuANg/arcgo/dbx/dialect/mysql"
	"github.com/DaiYuANg/arcgo/dbx/dialect/postgres"
	sqlitedialect "github.com/DaiYuANg/arcgo/dbx/dialect/sqlite"
	"github.com/DaiYuANg/arcgo/dbx/migrate"
	"github.com/DaiYuANg/arcgo/logx"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func loadConfig() (config2.AppConfig, error) {
	var cfg config2.AppConfig
	loader := configx.New(
		configx.WithTypedDefaults(config2.DefaultAppConfig()),
		configx.WithDotenv(".env"),
		configx.WithEnvPrefix("APP_"),
	)
	if err := loader.Load(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func selectDialect(driver string) (dialect.Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite":
		return sqlitedialect.Dialect{}, nil
	case "mysql":
		return mysql.Dialect{}, nil
	case "postgres", "postgresql":
		return postgres.Dialect{}, nil
	default:
		return nil, fmt.Errorf("unsupported db driver %q", driver)
	}
}

func main() {
	logger := logx.MustNew(logx.WithConsole(true), logx.WithTraceLevel())
	defer func() { _ = logx.Close(logger) }()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("load config failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	opts := db2.DefaultOpts(logger)
	if cfg.DB.NodeID != 0 {
		opts = append(opts, dbx.WithNodeID(cfg.DB.NodeID))
	}
	db, err := db2.Open(cfg.DB.Driver, cfg.DB.DSN, opts...)
	if err != nil {
		logger.Error("open database failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	d, err := selectDialect(cfg.DB.Driver)
	if err != nil {
		logger.Error("select dialect failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	migrateDir := os.Getenv("APP_MIGRATE_DIR")
	useEmbedded := migrateDir == ""
	if migrateDir == "" {
		migrateDir = "migrations"
	}

	runner := migrate.NewRunner(db.SQLDB(), d, migrate.RunnerOptions{
		HistoryTable:    "schema_migrations",
		AllowOutOfOrder: false,
		ValidateHash:    true,
	})

	ctx := context.Background()
	source := migrate.FileSource{
		FS:  embeddedMigrations,
		Dir: migrateDir,
	}
	if !useEmbedded {
		source = migrate.FileSource{
			FS:  os.DirFS("."),
			Dir: migrateDir,
		}
	}

	report, err := runner.UpSQL(ctx, migrate.FileSource{
		FS:  source.FS,
		Dir: source.Dir,
	})
	if err != nil {
		logger.Error("run migrations failed", slog.String("error", err.Error()), slog.String("dir", migrateDir))
		os.Exit(1)
	}

	logger.Info("migrations completed",
		slog.Int("applied", len(report.Applied)),
		slog.String("dir", migrateDir),
		slog.Bool("embedded", useEmbedded),
	)
}
