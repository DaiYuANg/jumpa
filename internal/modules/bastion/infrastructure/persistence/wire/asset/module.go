package asset

import (
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
	dbxrepo "github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion-persistence-asset",
	dix.WithModuleProviders(
		dix.Provider1(func(db *dbx.DB) ports.HostRepository { return dbxrepo.NewHostRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.HostAccountRepository { return dbxrepo.NewHostAccountRepository(db) }),
	),
)
