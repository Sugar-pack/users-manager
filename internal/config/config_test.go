package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAppConfig(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `api:
  bind: :9090
monitoring:
  rollback_timeout: 5s
db:
  conn_string: conn
  max_open_conns: 5
  conn_max_lifetime: 1s
  migration_dir_path: ./m
  migration_table: migrations`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	cfg, err := GetAppConfig()
	if err != nil {
		t.Fatalf("GetAppConfig: %v", err)
	}
	if cfg.API.Bind != ":9090" {
		t.Fatalf("unexpected bind: %s", cfg.API.Bind)
	}
	if cfg.Monitoring.RollbackTimeout.String() != "5s" {
		t.Fatalf("unexpected timeout: %v", cfg.Monitoring.RollbackTimeout)
	}
	if cfg.Db.ConnString != "conn" {
		t.Fatalf("unexpected conn string: %s", cfg.Db.ConnString)
	}
}
