package wire

import (
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
	dbxrepo "github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion-persistence",
	dix.WithModuleImports(db2.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(db *dbx.DB) ports.HostRepository { return dbxrepo.NewHostRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.HostAccountRepository { return dbxrepo.NewHostAccountRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.PolicyRepository { return dbxrepo.NewPolicyRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.PrincipalAccessRepository { return dbxrepo.NewPrincipalAccessRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.SessionRepository { return dbxrepo.NewSessionRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.SessionEventRepository { return dbxrepo.NewSessionEventRepository(db) }),
	),
)
