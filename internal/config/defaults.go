package config

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Server: struct {
			Port int `koanf:"port"`
		}{
			Port: 8080,
		},
		DB: struct {
			Driver string `koanf:"driver"`
			DSN    string `koanf:"dsn"`
			NodeID uint16 `koanf:"node_id"`
		}{
			Driver: "sqlite",
			DSN:    "file:backend?mode=memory&cache=shared",
			NodeID: 0,
		},
		Scheduler: struct {
			Enabled      bool `koanf:"enabled"`
			HeartbeatSec int  `koanf:"heartbeat_sec"`
			Distributed  struct {
				Enabled   bool   `koanf:"enabled"`
				KeyPrefix string `koanf:"key_prefix"`
				TTLSec    int    `koanf:"ttl_sec"`
			} `koanf:"distributed"`
		}{
			Enabled:      true,
			HeartbeatSec: 60,
			Distributed: struct {
				Enabled   bool   `koanf:"enabled"`
				KeyPrefix string `koanf:"key_prefix"`
				TTLSec    int    `koanf:"ttl_sec"`
			}{
				Enabled:   false,
				KeyPrefix: "gocron:lock",
				TTLSec:    30,
			},
		},
		Valkey: struct {
			Enabled  bool   `koanf:"enabled"`
			Addr     string `koanf:"addr"`
			Password string `koanf:"password"`
			DB       int    `koanf:"db"`
			UseTLS   bool   `koanf:"use_tls"`
		}{
			Enabled:  false,
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
			UseTLS:   false,
		},
		JWT: struct {
			Secret         string `koanf:"secret"`
			Issuer         string `koanf:"issuer"`
			AccessTTLMin   int    `koanf:"access_ttl_min"`
			RefreshTTLHour int    `koanf:"refresh_ttl_hour"`
		}{
			Secret:         "change-me-in-production",
			Issuer:         "arcgo-rbac-template",
			AccessTTLMin:   30,
			RefreshTTLHour: 168,
		},
		Authz: struct {
			ProtectedPrefix      string `koanf:"protected_prefix"`
			PublicPathsCSV       string `koanf:"public_paths_csv"`
			AuthOnlyResourcesCSV string `koanf:"auth_only_resources_csv"`
		}{
			ProtectedPrefix:      "/api",
			PublicPathsCSV:       "/api/auth/login,/api/auth/refresh,/api/health",
			AuthOnlyResourcesCSV: "me,auth",
		},
	}
}
