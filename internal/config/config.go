package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// DB contains database and migration settings.
type DB struct {
	ConnString       string        `mapstructure:"conn_string"`
	MigrationDirPath string        `mapstructure:"migration_dir_path"`
	MigrationTable   string        `mapstructure:"migration_table"`
	MaxOpenConns     int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime  time.Duration `mapstructure:"conn_max_lifetime"`
}

// API contains api settings.
type API struct {
	Bind string `mapstructure:"bind"`
}

type Monitoring struct {
	RollbackTimeout time.Duration `mapstructure:"rollback_timeout"`
}

// AppConfig is a container for application config.
type AppConfig struct {
	API        *API        `mapstructure:"api"`
	Db         *DB         `mapstructure:"db"`
	Monitoring *Monitoring `mapstructure:"monitoring"`
}

// GetAppConfig returns *Config.
func GetAppConfig() (*AppConfig, error) {
	viper.SetConfigName("config") // hardcoded config name
	viper.AddConfigPath(".")      // hardcoded configfile path
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read config from file: %w", err)
	}
	viper.AutomaticEnv()

	config := new(AppConfig)
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}

	return config, nil
}
