package audit

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/modules/audit/application"
	"github.com/DaiYuANg/jumpa/internal/modules/audit/infrastructure/persistence/wire"
	"github.com/DaiYuANg/jumpa/internal/modules/audit/ports"
)

var Module = dix.NewModule("audit",
	dix.WithModuleImports(wire.Module),
	dix.WithModuleProviders(
		dix.Provider1(func(repo ports.SessionEventRepository) application.SessionEventService {
			return application.NewSessionEventService(repo)
		}),
	),
)
