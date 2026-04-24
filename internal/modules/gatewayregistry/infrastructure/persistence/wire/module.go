package wire

import (
	"github.com/arcgolabs/dbx"
	"github.com/arcgolabs/dix"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
	dbxrepo "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/infrastructure/persistence/dbx"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/ports"
)

var Module = dix.NewModule("gateway-registry-persistence",
	dix.WithModuleImports(db2.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(db *dbx.DB) ports.GatewayRepository { return dbxrepo.NewGatewayRepository(db) }),
	),
)
