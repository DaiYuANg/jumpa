package pkg

import (
	"fmt"
	"strings"

	"github.com/arcgolabs/dbx/dialect"
	"github.com/arcgolabs/dbx/dialect/mysql"
	"github.com/arcgolabs/dbx/dialect/postgres"
	sqlitedialect "github.com/arcgolabs/dbx/dialect/sqlite"
)

func SelectDialect(driver string) (dialect.Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite":
		return sqlitedialect.New(), nil
	case "mysql", "mariadb":
		return mysql.New(), nil
	case "postgres", "postgresql":
		return postgres.New(), nil
	default:
		return nil, fmt.Errorf("unsupported db driver %q", driver)
	}
}
