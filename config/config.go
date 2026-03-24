package config

type AppConfig struct {
	Server struct {
		Port int `koanf:"port"`
	} `koanf:"server"`
	DB struct {
		Driver string `koanf:"driver"`
		DSN string `koanf:"dsn"`
	} `koanf:"db"`
}
