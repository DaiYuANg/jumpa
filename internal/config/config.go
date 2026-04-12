package config

type AppConfig struct {
	App struct {
		Name string `koanf:"name"`
	} `koanf:"app"`
	Server struct {
		Port int `koanf:"port"`
	} `koanf:"server"`
	DB struct {
		Driver string `koanf:"driver"`
		DSN    string `koanf:"dsn"`
		// Optional; if 0, dbx resolves node id from hostname.
		NodeID uint16 `koanf:"node_id"`
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
	Identity struct {
		Provider string `koanf:"provider"`
		OS       struct {
			Backend          string `koanf:"backend"`
			PAMService       string `koanf:"pam_service"`
			DirectoryNode    string `koanf:"directory_node"`
			CreateHomePolicy string `koanf:"create_home_policy"`
		} `koanf:"os"`
	} `koanf:"identity"`
	Bastion struct {
		Enabled bool `koanf:"enabled"`
		SSH     struct {
			ListenAddr       string `koanf:"listen_addr"`
			HostKeyPath      string `koanf:"host_key_path"`
			HostKeyPolicy    string `koanf:"host_key_policy"`
			KnownHostsPath   string `koanf:"known_hosts_path"`
			TrustedProxyCIDR string `koanf:"trusted_proxy_cidr"`
		} `koanf:"ssh"`
		Session struct {
			IdleTimeoutMin     int    `koanf:"idle_timeout_min"`
			MaxDurationMin     int    `koanf:"max_duration_min"`
			RecordingDirectory string `koanf:"recording_directory"`
		} `koanf:"session"`
		Access struct {
			ApprovalTTLMin int `koanf:"approval_ttl_min"`
		} `koanf:"access"`
	} `koanf:"bastion"`
	Gateway struct {
		Registry struct {
			NodeName        string `koanf:"node_name"`
			NodeKey         string `koanf:"node_key"`
			AdvertiseAddr   string `koanf:"advertise_addr"`
			Zone            string `koanf:"zone"`
			TagsCSV         string `koanf:"tags_csv"`
			HeartbeatSec    int    `koanf:"heartbeat_sec"`
			OfflineAfterSec int    `koanf:"offline_after_sec"`
		} `koanf:"registry"`
	} `koanf:"gateway"`
	Audit struct {
		StoreCommandInput bool `koanf:"store_command_input"`
		StoreReplayStream bool `koanf:"store_replay_stream"`
	} `koanf:"audit"`
	Authz struct {
		ProtectedPrefix      string `koanf:"protected_prefix"`
		PublicPathsCSV       string `koanf:"public_paths_csv"`
		AuthOnlyResourcesCSV string `koanf:"auth_only_resources_csv"`
	} `koanf:"authz"`
}
