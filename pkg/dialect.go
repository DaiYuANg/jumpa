package pkg

import (
	"fmt"
	"strings"

	"github.com/DaiYuANg/arcgo/dbx/dialect"
	"github.com/DaiYuANg/arcgo/dbx/dialect/mysql"
	"github.com/DaiYuANg/arcgo/dbx/dialect/postgres"
	sqlitedialect "github.com/DaiYuANg/arcgo/dbx/dialect/sqlite"
)

func SelectDialect(driver string) (dialect.Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite":
		return sqlitedialect.New(), nil
	case "mysql":
		return mysql.New(), nil
	case "postgres", "postgresql":
		return postgres.New(), nil
	default:
		return nil, fmt.Errorf("unsupported db driver %q", driver)
	}
}
