package wire

import (
	"github.com/DaiYuANg/arcgo/dix"
	db2 "github.com/DaiYuANg/jumpa/internal/db"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/wire/access"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/wire/asset"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/wire/session"
)

var Module = dix.NewModule("bastion-persistence",
	dix.WithModuleImports(db2.Module, asset.Module, access.Module, session.Module),
)
