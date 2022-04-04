package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// DB contains database and migration settings
type DB struct {
	ConnString       string        `mapstructure:"conn_string"`
	MaxOpenConns     int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime  time.Duration `mapstructure:"conn_max_lifetime"`
	MigrationDirPath string        `mapstructure:"migration_dir_path"`
	MigrationTable   string        `mapstructure:"migration_table"`
}

// API contains api settings
type API struct {
	Bind string `mapstructure:"bind"`
}

// AppConfig is a container for application config
type AppConfig struct {
	API *API `mapstructure:"api"`
	Db  *DB  `mapstructure:"db"`
}

// GetAppConfig returns *Config
func GetAppConfig() (*AppConfig, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read config from file: %w", err)
	}
	viper.AutomaticEnv()

	config := new(AppConfig)
	err := viper.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return config, nil
}
