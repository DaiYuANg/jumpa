package config

func DefaultAppConfig() AppConfig {
	return AppConfig{
		App: struct {
			Name string `koanf:"name"`
		}{
			Name: "jumpa",
		},
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
			Issuer:         "jumpa",
			AccessTTLMin:   30,
			RefreshTTLHour: 168,
		},
		Identity: struct {
			Provider string `koanf:"provider"`
			OS       struct {
				Backend          string `koanf:"backend"`
				PAMService       string `koanf:"pam_service"`
				DirectoryNode    string `koanf:"directory_node"`
				CreateHomePolicy string `koanf:"create_home_policy"`
			} `koanf:"os"`
		}{
			Provider: "local",
			OS: struct {
				Backend          string `koanf:"backend"`
				PAMService       string `koanf:"pam_service"`
				DirectoryNode    string `koanf:"directory_node"`
				CreateHomePolicy string `koanf:"create_home_policy"`
			}{
				Backend:          "auto",
				PAMService:       "sshd",
				DirectoryNode:    "/Search",
				CreateHomePolicy: "manual",
			},
		},
		Bastion: struct {
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
		}{
			Enabled: true,
			SSH: struct {
				ListenAddr       string `koanf:"listen_addr"`
				HostKeyPath      string `koanf:"host_key_path"`
				HostKeyPolicy    string `koanf:"host_key_policy"`
				KnownHostsPath   string `koanf:"known_hosts_path"`
				TrustedProxyCIDR string `koanf:"trusted_proxy_cidr"`
			}{
				ListenAddr:       ":2222",
				HostKeyPath:      "./secrets/ssh_host_ed25519_key",
				HostKeyPolicy:    "insecure",
				KnownHostsPath:   "./secrets/known_hosts",
				TrustedProxyCIDR: "",
			},
			Session: struct {
				IdleTimeoutMin     int    `koanf:"idle_timeout_min"`
				MaxDurationMin     int    `koanf:"max_duration_min"`
				RecordingDirectory string `koanf:"recording_directory"`
			}{
				IdleTimeoutMin:     15,
				MaxDurationMin:     480,
				RecordingDirectory: "./data/recordings",
			},
			Access: struct {
				ApprovalTTLMin int `koanf:"approval_ttl_min"`
			}{
				ApprovalTTLMin: 30,
			},
		},
		Gateway: struct {
			Registry struct {
				NodeName        string `koanf:"node_name"`
				NodeKey         string `koanf:"node_key"`
				AdvertiseAddr   string `koanf:"advertise_addr"`
				Zone            string `koanf:"zone"`
				TagsCSV         string `koanf:"tags_csv"`
				HeartbeatSec    int    `koanf:"heartbeat_sec"`
				OfflineAfterSec int    `koanf:"offline_after_sec"`
			} `koanf:"registry"`
		}{
			Registry: struct {
				NodeName        string `koanf:"node_name"`
				NodeKey         string `koanf:"node_key"`
				AdvertiseAddr   string `koanf:"advertise_addr"`
				Zone            string `koanf:"zone"`
				TagsCSV         string `koanf:"tags_csv"`
				HeartbeatSec    int    `koanf:"heartbeat_sec"`
				OfflineAfterSec int    `koanf:"offline_after_sec"`
			}{
				NodeName:        "",
				NodeKey:         "",
				AdvertiseAddr:   "",
				Zone:            "default",
				TagsCSV:         "ssh",
				HeartbeatSec:    15,
				OfflineAfterSec: 60,
			},
		},
		Audit: struct {
			StoreCommandInput bool `koanf:"store_command_input"`
			StoreReplayStream bool `koanf:"store_replay_stream"`
		}{
			StoreCommandInput: true,
			StoreReplayStream: true,
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
