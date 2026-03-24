package config

type AppConfig struct {
	Server struct {
		Port int `koanf:"port"`
	} `koanf:"server"`
	DB struct {
		Driver string `koanf:"driver"`
		DSN    string `koanf:"dsn"`
	} `koanf:"db"`
	Scheduler struct {
		Enabled      bool `koanf:"enabled"`
		HeartbeatSec int  `koanf:"heartbeat_sec"`
		Distributed  struct {
			Enabled   bool   `koanf:"enabled"`
			KeyPrefix string `koanf:"key_prefix"`
			TTLSec    int    `koanf:"ttl_sec"`
		} `koanf:"distributed"`
	} `koanf:"scheduler"`
	Valkey struct {
		Enabled  bool   `koanf:"enabled"`
		Addr     string `koanf:"addr"`
		Password string `koanf:"password"`
		DB       int    `koanf:"db"`
		UseTLS   bool   `koanf:"use_tls"`
	} `koanf:"valkey"`
	JWT struct {
		Secret         string `koanf:"secret"`
		Issuer         string `koanf:"issuer"`
		AccessTTLMin   int    `koanf:"access_ttl_min"`
		RefreshTTLHour int    `koanf:"refresh_ttl_hour"`
	} `koanf:"jwt"`
	Authz struct {
		ProtectedPrefix     string `koanf:"protected_prefix"`
		PublicPathsCSV      string `koanf:"public_paths_csv"`
		AuthOnlyResourcesCSV string `koanf:"auth_only_resources_csv"`
	} `koanf:"authz"`
}
