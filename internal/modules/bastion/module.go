package bastion

import (
	"github.com/arcgolabs/dix"
	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/access"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/asset"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/infrastructure/persistence/wire"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/overview"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/session"
)

var Module = dix.NewModule("bastion",
	dix.WithModuleImports(
		config2.Module,
		identity.Module,
		wire.Module,
		overview.Module,
		asset.Module,
		access.Module,
		session.Module,
	),
)
