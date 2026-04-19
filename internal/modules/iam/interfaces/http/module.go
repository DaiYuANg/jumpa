package http

import (
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/httpendpoint"
	"github.com/DaiYuANg/jumpa/internal/modules/iam"
)

var Module = dix.NewModule("iam-http",
	dix.WithModuleImports(iam.Module),
	dix.WithModuleProviders(
		httpendpoint.Provider3("iam.users", NewUserEndpoint),
		httpendpoint.Provider3("iam.rbac", NewRBACEndpoint),
	),
)
