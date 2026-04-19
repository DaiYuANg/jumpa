package main

import (
	"context"
	"embed"
	"log/slog"
	"os"

	"github.com/DaiYuANg/arcgo/configx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/migrate"
	"github.com/DaiYuANg/arcgo/logx"
	"github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/db"
	"github.com/DaiYuANg/jumpa/pkg"
)

const (
	migrateRuntimeName = "jumpa-migrate"
	defaultMigrateDir  = "migrations"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func runMigrate() error {
	logger := newMigrateLogger()
	defer func() { _ = logx.Close(logger) }()

	cfg, err := loadMigrateConfig()
	if err != nil {
		logger.Error("load config failed", slog.String("error", err.Error()))
		return err
	}

	sqlDB, err := openMigrateDB(cfg, logger)
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	report, dir, embedded, err := executeMigrations(sqlDB, cfg)
	if err != nil {
		logger.Error("run migrations failed", slog.String("error", err.Error()), slog.String("dir", dir))
		return err
	}

	logger.Info("migrations completed",
		slog.String("runtime", migrateRuntimeName),
		slog.Int("applied", report.Applied.Len()),
		slog.String("dir", dir),
		slog.Bool("embedded", embedded),
	)

	return nil
}

func newMigrateLogger() *slog.Logger {
	return logx.MustNew(logx.WithConsole(true), logx.WithDebugLevel())
}

func loadMigrateConfig() (config.AppConfig, error) {
	var cfg config.AppConfig
	loader := configx.New(
		configx.WithTypedDefaults(config.DefaultAppConfig()),
		configx.WithDotenv(".env"),
		configx.WithEnvPrefix("APP_"),
	)
	if err := loader.Load(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func openMigrateDB(cfg config.AppConfig, logger *slog.Logger) (*dbx.DB, error) {
	opts := db.DefaultOpts(logger)
	if cfg.DB.NodeID != 0 {
		opts = append(opts, dbx.WithNodeID(cfg.DB.NodeID))
	}

	database, err := db.Open(cfg.DB.Driver, cfg.DB.DSN, opts...)
	if err != nil {
		logger.Error("open database failed", slog.String("error", err.Error()))
		return nil, err
	}

	return database, nil
}

func executeMigrations(database *dbx.DB, cfg config.AppConfig) (migrate.RunReport, string, bool, error) {
	dialect, err := pkg.SelectDialect(cfg.DB.Driver)
	if err != nil {
		return migrate.RunReport{}, "", false, err
	}

	dir, embedded := resolveMigrationDir()
	source := migrationSource(dir, embedded)
	runner := migrate.NewRunner(database.SQLDB(), dialect, migrate.RunnerOptions{
		HistoryTable:    "schema_migrations",
		AllowOutOfOrder: false,
		ValidateHash:    true,
	})

	report, err := runner.UpSQL(context.Background(), source)
	if err != nil {
		return migrate.RunReport{}, dir, embedded, err
	}

	return report, dir, embedded, nil
}

func resolveMigrationDir() (string, bool) {
	dir := os.Getenv("APP_MIGRATE_DIR")
	if dir == "" {
		return defaultMigrateDir, true
	}
	return dir, false
}

func migrationSource(dir string, embedded bool) migrate.FileSource {
	if embedded {
		return migrate.FileSource{
			FS:  embeddedMigrations,
			Dir: dir,
		}
	}

	return migrate.FileSource{
		FS:  os.DirFS("."),
		Dir: dir,
	}
}
