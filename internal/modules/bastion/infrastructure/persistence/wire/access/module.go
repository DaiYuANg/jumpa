package access

import (
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dix"
	dbxrepo "github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

var Module = dix.NewModule("bastion-persistence-access",
	dix.WithModuleProviders(
		dix.Provider1(func(db *dbx.DB) ports.PolicyRepository { return dbxrepo.NewPolicyRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.PrincipalAccessRepository { return dbxrepo.NewPrincipalAccessRepository(db) }),
		dix.Provider1(func(db *dbx.DB) ports.AccessRequestRepository { return dbxrepo.NewAccessRequestRepository(db) }),
	),
)
