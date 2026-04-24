package wire

import (
	"github.com/arcgolabs/dbx"
	"github.com/arcgolabs/dix"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
	dbxrepo "github.com/DaiYuANg/jumpa/internal/modules/audit/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/audit/ports"
)

var Module = dix.NewModule("audit-persistence",
	dix.WithModuleImports(db2.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(db *dbx.DB) ports.SessionEventRepository { return dbxrepo.NewSessionEventRepository(db) }),
	),
)
