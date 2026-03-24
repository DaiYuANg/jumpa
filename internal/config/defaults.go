package config

func DefaultAppConfig() AppConfig {
	var cfg AppConfig
	cfg.Server.Port = 8080
	cfg.DB.Driver = "sqlite"
	cfg.DB.DSN = "file:backend?mode=memory&cache=shared"
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.HeartbeatSec = 60
	cfg.Scheduler.Distributed.Enabled = false
	cfg.Scheduler.Distributed.KeyPrefix = "gocron:lock"
	cfg.Scheduler.Distributed.TTLSec = 30
	cfg.Valkey.Enabled = false
	cfg.Valkey.Addr = "127.0.0.1:6379"
	cfg.Valkey.Password = ""
	cfg.Valkey.DB = 0
	cfg.Valkey.UseTLS = false
	cfg.JWT.Secret = "change-me-in-production"
	cfg.JWT.Issuer = "arcgo-rbac-template"
	cfg.JWT.AccessTTLMin = 30
	cfg.JWT.RefreshTTLHour = 168
	cfg.Authz.ProtectedPrefix = "/api"
	cfg.Authz.PublicPathsCSV = "/api/auth/login,/api/auth/refresh,/api/health"
	cfg.Authz.AuthOnlyResourcesCSV = "me,auth"
	return cfg
}
